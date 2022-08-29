package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFERegistryGPGKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFERegistryGPGKeyCreate,
		Read:   resourceTFERegistryGPGKeyRead,
		Update: resourceTFERegistryGPGKeyUpdate,
		Delete: resourceTFERegistryGPGKeyDelete,

		Schema: map[string]*schema.Schema{
			"provider_namespace": {
				Type:     schema.TypeString,
				Required: true,
			},

			"ascii_armor": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTFERegistryGPGKeyCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	providerNamespace := d.Get("provider_namespace").(string)
	asciiArmor := d.Get("ascii_armor").(string)
	gpgKeyType := d.Get("type").(string)

	options := tfe.GPGKeyCreateOptions{
		Type:       gpgKeyType,
		Namespace:  providerNamespace,
		AsciiArmor: asciiArmor,
	}

	registryName := tfe.RegistryName("private")

	log.Printf("[DEBUG] Create new GPG key for namespace/organization %s", providerNamespace)
	gpgKey, err := tfeClient.GPGKeys.Create(ctx, registryName, options)

	if err != nil {
		return fmt.Errorf(
			"Error creating new GPG key for namespace/organization %s: %w", providerNamespace, err)
	}

	// The key ID needs to be set at creation because it is computed and is required for successful reads
	d.Set("key_id", gpgKey.KeyID)

	d.SetId(gpgKey.ID)

	return resourceTFERegistryGPGKeyRead(d, meta)
}

func resourceTFERegistryGPGKeyRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	gpgKeyID := tfe.GPGKeyID{
		RegistryName: tfe.RegistryName("private"),
		Namespace:    d.Get("provider_namespace").(string),
		KeyID:        d.Get("key_id").(string),
	}

	log.Printf("[DEBUG] Read GPG key %s", d.Id())
	gpgKey, err := tfeClient.GPGKeys.Read(ctx, gpgKeyID)

	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] GPG key %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error reading GPG key %s: %w", d.Id(), err)
	}

	d.Set("provider_namespace", gpgKey.Namespace)
	d.Set("ascii_armor", gpgKey.AsciiArmor)
	d.Set("type", "gpg-keys") // This shouldn't be hard-coded, but I don't know what else to do because go-tfe doesn't return the type
	d.Set("key_id", gpgKey.KeyID)

	return nil
}

func resourceTFERegistryGPGKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	gpgKeyID := tfe.GPGKeyID{
		RegistryName: tfe.RegistryName("private"),
		Namespace:    d.Get("provider_namespace").(string),
		KeyID:        d.Get("key_id").(string),
	}

	options := tfe.GPGKeyUpdateOptions{
		Type:      d.Get("type").(string),
		Namespace: d.Get("namespace").(string),
	}

	log.Printf("[DEBUG] Update GPG key: %s", d.Id())
	_, err := tfeClient.GPGKeys.Update(ctx, gpgKeyID, options)

	if err != nil {
		return fmt.Errorf("Error updating GPG key %s: %w", d.Id(), err)
	}

	return resourceTFERegistryGPGKeyRead(d, meta)
}

func resourceTFERegistryGPGKeyDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	gpgKeyID := tfe.GPGKeyID{
		RegistryName: tfe.RegistryName("private"),
		Namespace:    d.Get("provider_namespace").(string),
		KeyID:        d.Get("key_id").(string),
	}

	log.Printf("[DEBUG] Delete GPG key: %s", d.Id())
	err := tfeClient.GPGKeys.Delete(ctx, gpgKeyID)

	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}

		return fmt.Errorf("Error deleting GPG key %s: %w", d.Id(), err)
	}

	return nil
}
