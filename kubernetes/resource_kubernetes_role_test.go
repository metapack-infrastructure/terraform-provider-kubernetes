package kubernetes

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	api "k8s.io/api/rbac/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAccKubernetesRole_basic(t *testing.T) {
	var conf api.Role
	name := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "kubernetes_role.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckKubernetesRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesRoleConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubernetesRoleExists("kubernetes_role.test", &conf),
					resource.TestCheckResourceAttr("kubernetes_role.test", "rule.#", "1"),
					resource.TestCheckResourceAttr("kubernetes_role.test", "rule.0.resources.#", "2"),
					resource.TestCheckResourceAttr("kubernetes_role.test", "rule.0.resources.0", "pods"),
					resource.TestCheckResourceAttr("kubernetes_role.test", "rule.0.resources.1", "pods/log"),
					resource.TestCheckResourceAttr("kubernetes_role.test", "rule.0.verbs.#", "2"),
					resource.TestCheckResourceAttr("kubernetes_role.test", "rule.0.verbs.0", "get"),
					resource.TestCheckResourceAttr("kubernetes_role.test", "rule.0.verbs.1", "list"),
				),
			},
			{
				Config: testAccKubernetesRoleConfig_modified(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubernetesRoleExists("kubernetes_role.test", &conf),
					resource.TestCheckResourceAttr("kubernetes_role.test", "rule.#", "2"),
					resource.TestCheckResourceAttr("kubernetes_role.test", "rule.0.verbs.#", "3"),
					resource.TestCheckResourceAttr("kubernetes_role.test", "rule.0.verbs.2", "watch"),
					resource.TestCheckResourceAttr("kubernetes_role.test", "rule.1.api_groups.#", "1"),
					resource.TestCheckResourceAttr("kubernetes_role.test", "rule.1.resources.#", "1"),
					resource.TestCheckResourceAttr("kubernetes_role.test", "rule.1.resources.0", "deployments"),
					resource.TestCheckResourceAttr("kubernetes_role.test", "rule.1.verbs.#", "2"),
					resource.TestCheckResourceAttr("kubernetes_role.test", "rule.1.verbs.0", "get"),
					resource.TestCheckResourceAttr("kubernetes_role.test", "rule.1.verbs.1", "list"),
				),
			},
		},
	})
}

func TestAccKubernetesRole_importBasic(t *testing.T) {
	resourceName := "kubernetes_role.test"
	name := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesRoleConfig_basic(name),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func testAccCheckKubernetesRoleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*kubernetesProvider).conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "kubernetes_role" {
			continue
		}
		namespace, name, err := idParts(rs.Primary.ID)
		if err != nil {
			return err
		}
		resp, err := conn.RbacV1().Roles(namespace).Get(name, meta_v1.GetOptions{})
		if err == nil {
			if resp.Name == rs.Primary.ID {
				return fmt.Errorf(" Role still exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckKubernetesRoleExists(n string, obj *api.Role) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*kubernetesProvider).conn
		namespace, name, err := idParts(rs.Primary.ID)
		if err != nil {
			return err
		}
		out, err := conn.RbacV1().Roles(namespace).Get(name, meta_v1.GetOptions{})
		if err != nil {
			return err
		}

		*obj = *out
		return nil
	}
}

func testAccKubernetesRoleConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "kubernetes_role" "test" {
	metadata {
		name = "%s"
	}
	rule {
		api_groups = [""]
		resources  = ["pods", "pods/log"]
		verbs = ["get", "list"]
	}
}`, name)
}

func testAccKubernetesRoleConfig_modified(name string) string {
	return fmt.Sprintf(`
resource "kubernetes_role" "test" {
	metadata {
		name = "%s"
	}
	rule {
		api_groups = [""]
		resources  = ["pods", "pods/log"]
		verbs      = ["get", "list", "watch"]
	}
	rule {
		api_groups = [""]
		resources  = ["deployments"]
		verbs      = ["get", "list"]
	}
}`, name)
}
