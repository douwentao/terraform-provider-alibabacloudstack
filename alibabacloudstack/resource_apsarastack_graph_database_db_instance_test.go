package alibabacloudstack

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/PaesslerAG/jsonpath"
	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"

	"github.com/aliyun/terraform-provider-alibabacloudstack/alibabacloudstack/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers(
		"alibabacloudstack_graph_database_db_instance",
		&resource.Sweeper{
			Name: "alibabacloudstack_graph_database_db_instance",
			F:    testSweepGraphDatabaseDbInstance,
		})
}

func testSweepGraphDatabaseDbInstance(region string) error {
	rawClient, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting Alicloud client: %s", err)
	}
	client := rawClient.(*connectivity.AlibabacloudStackClient)
	prefixes := []string{
		"tf-testAcc",
		"tf_testAcc",
	}
	action := "DescribeDBInstances"
	request := map[string]interface{}{}

	request["RegionId"] = client.RegionId
	request["PageSize"] = PageSizeLarge
	request["PageNumber"] = 1

	request["product"] = "gdb"
	request["OrganizationId"] = client.Department
	var response map[string]interface{}
	conn, err := client.NewGdbClient()
	if err != nil {
		log.Printf("[ERROR] %s get an error: %#v", action, err)
	}
	for {
		runtime := util.RuntimeOptions{IgnoreSSL: tea.Bool(client.Config.Insecure)}
		runtime.SetAutoretry(true)
		wait := incrementalWait(3*time.Second, 3*time.Second)
		err = resource.Retry(5*time.Minute, func() *resource.RetryError {
			response, err = conn.DoRequest(StringPointer(action), nil, StringPointer("POST"), StringPointer("2019-09-03"), StringPointer("AK"), nil, request, &runtime)
			if err != nil {
				if NeedRetry(err) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		addDebug(action, response, request)
		if err != nil {
			log.Printf("[ERROR] %s get an error: %#v", action, err)
			return nil
		}

		resp, err := jsonpath.Get("$.Items.DBInstance", response)
		if err != nil {
			log.Printf("[ERROR] Getting resource %s attribute by path %s failed!!! Body: %v.", "$.Items.DBInstance", action, err)
			return nil
		}
		result, _ := resp.([]interface{})
		for _, v := range result {
			item := v.(map[string]interface{})

			skip := true
			for _, prefix := range prefixes {
				if strings.HasPrefix(strings.ToLower(item["DBInstanceDescription"].(string)), strings.ToLower(prefix)) {
					skip = false
				}
			}
			if item["DBInstanceStatus"].(string) != "Running" {
				skip = true
			}
			if skip {
				log.Printf("[INFO] Skipping Graph Database DbInstance: %s", item["DBInstanceDescription"].(string))
				continue
			}
			action := "DeleteDBInstance"
			deleteRequest := map[string]interface{}{
				"DBInstanceId": item["DBInstanceId"],
			}
			_, err = conn.DoRequest(StringPointer(action), nil, StringPointer("POST"), StringPointer("2019-09-03"), StringPointer("AK"), nil, deleteRequest, &util.RuntimeOptions{IgnoreSSL: tea.Bool(client.Config.Insecure)})
			if err != nil {
				log.Printf("[ERROR] Failed to delete Graph Database DbInstance (%s): %s", item["DBInstanceDescription"].(string), err)
			}
			log.Printf("[INFO] Delete Graph Database DbInstance success: %s ", item["DBInstanceDescription"].(string))
		}
		if len(result) < PageSizeLarge {
			break
		}
		request["PageNumber"] = request["PageNumber"].(int) + 1
	}
	return nil
}

func TestAccAlibabacloudStackGraphDatabaseDbInstance_basic0(t *testing.T) {
	var v map[string]interface{}
	resourceId := "alibabacloudstack_graph_database_db_instance.default"
	ra := resourceAttrInit(resourceId, AlibabacloudStackGraphDatabaseDbInstanceMap0)
	rc := resourceCheckInitWithDescribeMethod(resourceId, &v, func() interface{} {
		return &GdbService{testAccProvider.Meta().(*connectivity.AlibabacloudStackClient)}
	}, "DescribeGraphDatabaseDbInstance")
	rac := resourceAttrCheckInit(rc, ra)
	testAccCheck := rac.resourceAttrMapUpdateSet()
	rand := acctest.RandIntRange(10000, 99999)
	name := fmt.Sprintf("tf-testacc%sgraphdatabasedbinstance%d", defaultRegionToTest, rand)
	testAccConfig := resourceTestAccConfigFunc(resourceId, name, AlibabacloudStackGraphDatabaseDbInstanceBasicDependence0)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		IDRefreshName: resourceId,
		Providers:     testAccProviders,
		CheckDestroy:  rac.checkResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccConfig(map[string]interface{}{
					"db_node_class":            "gdb.r.xlarge",
					"db_instance_network_type": "vpc",
					"db_version":               "1.0",
					"db_instance_category":     "HA",
					"db_instance_storage_type": "cloud_essd",
					"db_node_storage":          "50",
					"payment_type":             "PayAsYouGo",
					"db_instance_description":  "${var.name}",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"db_node_class":            "gdb.r.xlarge",
						"db_instance_network_type": "vpc",
						"db_version":               "1.0",
						"db_instance_category":     "HA",
						"db_instance_storage_type": "cloud_essd",
						"db_node_storage":          "50",
						"payment_type":             "PayAsYouGo",
						"db_instance_description":  name,
					}),
				),
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"db_instance_ip_array": []map[string]interface{}{
						{
							"db_instance_ip_array_name": "default",
							"security_ips":              "127.0.0.2",
						},
						{
							"db_instance_ip_array_name": "tftest",
							"security_ips":              "192.168.0.1",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"db_instance_ip_array.#": "2",
					}),
				),
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"db_instance_description": "${var.name}_update",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"db_instance_description": name + "_update",
					}),
				),
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"db_instance_description": "${var.name}",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"db_instance_description": name,
					}),
				),
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"db_node_storage": "80",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"db_node_storage": "80",
					}),
				),
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"db_node_class": "gdb.r.2xlarge",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"db_node_class": "gdb.r.2xlarge",
					}),
				),
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"db_instance_description": "${var.name}",
					"db_node_storage":         "100",
					"db_node_class":           "gdb.r.xlarge",
					"db_instance_ip_array": []map[string]interface{}{
						{
							"db_instance_ip_array_name": "default",
							"security_ips":              "127.0.0.1",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"db_instance_description": name,
						"db_node_storage":         "100",
						"db_node_class":           "gdb.r.xlarge",
						"db_instance_ip_array.#":  "1",
					}),
				),
			},
			{
				ResourceName:            resourceId,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auto_renew", "auto_renew_period", "order_param", "period", "engine_version", "effective_time", "used_time", "security_ip_list", "order_type", "maintain_time", "zone_id"},
			},
		},
	})
}

var AlibabacloudStackGraphDatabaseDbInstanceMap0 = map[string]string{
	"engine_version":    NOSET,
	"period":            NOSET,
	"effective_time":    NOSET,
	"used_time":         NOSET,
	"order_type":        NOSET,
	"security_ip_list":  NOSET,
	"auto_renew":        NOSET,
	"order_param":       NOSET,
	"auto_renew_period": NOSET,
}

func AlibabacloudStackGraphDatabaseDbInstanceBasicDependence0(name string) string {
	return fmt.Sprintf(` 
variable "name" {
  default = "%s"
}
`, name)
}
