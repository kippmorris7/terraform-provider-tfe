package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFERegistryGPGKeyDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryGPGKeyDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_registry_gpg_key.foobar", "provider_namespace", orgName),
					resource.TestCheckResourceAttr(
						"data.tfe_registry_gpg_key.foobar", "ascii_armor", sampleAsciiArmor+"\n"),
					resource.TestCheckResourceAttrSet("data.tfe_registry_gpg_key.foobar", "id"),
				),
			},
		},
	})
}

func testAccTFERegistryGPGKeyDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_registry_gpg_key" "foobar" {
	provider_namespace = tfe_organization.foobar.name
	ascii_armor        = <<EOT
%s
EOT
}

data "tfe_registry_gpg_key" "foobar" {
  provider_namespace = tfe_registry_gpg_key.foobar.provider_namespace
  key_id             = tfe_registry_gpg_key.foobar.key_id
}`, rInt, sampleAsciiArmor)
}
