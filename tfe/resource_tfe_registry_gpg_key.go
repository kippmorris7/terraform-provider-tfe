package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Since these are the only possible values (respectively) for these two attributes,
// it doesn't make sense to expose them to users of the provider.
var privateRegistryName = tfe.RegistryName("private")
var gpgKeyType = "gpg-keys"

func resourceTFERegistryGPGKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFERegistryGPGKeyCreate,
		Read:   resourceTFERegistryGPGKeyRead,
		Delete: resourceTFERegistryGPGKeyDelete,

		Schema: map[string]*schema.Schema{
			"provider_namespace": {
				Type:     schema.TypeString,
				Required: true,
				// go-tfe and the HTTP API support updating the namespace, but because the namespace
				// is part of what uniquely identifies a key, per the API structure and the GPGKeyID
				// type from go-tfe, modifying the namespace must force a new resource in Terraform.
				ForceNew: true,
			},

			"ascii_armor": {
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

	options := tfe.GPGKeyCreateOptions{
		Type:       gpgKeyType,
		Namespace:  providerNamespace,
		AsciiArmor: asciiArmor,
	}

	log.Printf("[DEBUG] Create new GPG key for namespace/organization %s", providerNamespace)
	gpgKey, err := tfeClient.GPGKeys.Create(ctx, privateRegistryName, options)

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
		RegistryName: privateRegistryName,
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
	d.Set("key_id", gpgKey.KeyID)

	return nil
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
