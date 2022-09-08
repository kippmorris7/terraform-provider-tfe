package tfe

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFERegistryGPGKey() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFERegistryGPGKeyRead,

		Schema: map[string]*schema.Schema{
			"provider_namespace": {
				Type:     schema.TypeString,
				Required: true,
			},

			"key_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"ascii_armor": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTFERegistryGPGKeyRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	providerNamespace := d.Get("provider_namespace").(string)
	keyId := d.Get("key_id").(string)

	gpgKeyID := tfe.GPGKeyID{
		RegistryName: privateRegistryName,
		Namespace:    providerNamespace,
		KeyID:        keyId,
	}

	gpgKey, err := tfeClient.GPGKeys.Read(ctx, gpgKeyID)

	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return fmt.Errorf("Could not find GPG key %s/%s", providerNamespace, keyId)
		}

		return fmt.Errorf("Error retrieving GPG key %s/%s", providerNamespace, keyId)
	}

	d.Set("provider_namespace", gpgKey.Namespace)
	d.Set("ascii_armor", gpgKey.AsciiArmor)
	d.Set("key_id", gpgKey.KeyID)
	d.SetId(gpgKey.ID)

	return nil
}
