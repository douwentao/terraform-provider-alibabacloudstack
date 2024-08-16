package alibabacloudstack

import (
	"fmt"
	"log"
	"time"

	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/terraform-provider-alibabacloudstack/alibabacloudstack/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAlibabacloudStackCloudFirewallControlPolicyOrder() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlibabacloudStackCloudFirewallControlPolicyOrderCreate,
		Read:   resourceAlibabacloudStackCloudFirewallControlPolicyOrderRead,
		Update: resourceAlibabacloudStackCloudFirewallControlPolicyOrderUpdate,
		Delete: resourceAlibabacloudStackCloudFirewallControlPolicyOrderDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"acl_uuid": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"direction": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"in", "out"}, false),
			},
			"order": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func resourceAlibabacloudStackCloudFirewallControlPolicyOrderCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AlibabacloudStackClient)
	var response map[string]interface{}
	action := "ModifyControlPolicyPriority"
	request := make(map[string]interface{})
	conn, err := client.NewCloudfwClient()
	if err != nil {
		return WrapError(err)
	}
	request["Direction"] = d.Get("direction")
	request["Order"] = d.Get("order")
	request["AclUuid"] = d.Get("acl_uuid")
	request["RegionId"] = client.RegionId
	request["Product"] = "Cloudfw"
	request["OrganizationId"] = client.Department
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err = conn.DoRequest(StringPointer(action), nil, StringPointer("POST"), StringPointer("2017-12-07"), StringPointer("AK"), nil, request, &util.RuntimeOptions{IgnoreSSL: tea.Bool(client.Config.Insecure)})
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
		return WrapErrorf(err, DefaultErrorMsg, "alibabacloudstack_cloud_firewall_control_policy_order", action, AlibabacloudStackSdkGoERROR)
	}

	d.SetId(fmt.Sprint(request["AclUuid"], ":", request["Direction"]))

	return resourceAlibabacloudStackCloudFirewallControlPolicyRead(d, meta)
}

func resourceAlibabacloudStackCloudFirewallControlPolicyOrderUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AlibabacloudStackClient)
	var response map[string]interface{}
	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}
	action := "ModifyControlPolicyPriority"
	conn, err := client.NewCloudfwClient()
	if err != nil {
		return WrapError(err)
	}

	update := false
	request := map[string]interface{}{
		"AclUuid":   parts[0],
		"Direction": parts[1],
	}
	request["RegionId"] = client.RegionId
	request["Product"] = "Cloudfw"
	request["OrganizationId"] = client.Department
	if d.HasChange("order") {
		update = true
		request["Order"] = d.Get("order")
	}

	if update {
		wait := incrementalWait(3*time.Second, 3*time.Second)
		err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			response, err = conn.DoRequest(StringPointer(action), nil, StringPointer("POST"), StringPointer("2017-12-07"), StringPointer("AK"), nil, request, &util.RuntimeOptions{IgnoreSSL: tea.Bool(client.Config.Insecure)})
			if err != nil {
				if NeedRetry(err) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
	}
	addDebug(action, response, request)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alibabacloudstack_cloud_firewall_control_policy_order", action, AlibabacloudStackSdkGoERROR)
	}

	d.SetId(fmt.Sprint(request["AclUuid"], ":", request["Direction"]))

	return resourceAlibabacloudStackCloudFirewallControlPolicyRead(d, meta)
}

func resourceAlibabacloudStackCloudFirewallControlPolicyOrderRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AlibabacloudStackClient)
	cloudfwService := CloudfwService{client}
	object, err := cloudfwService.DescribeCloudFirewallControlPolicy(d.Id())
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[DEBUG] Resource alibabacloudstack_cloud_firewall_control_policy_order cloudfwService.DescribeCloudFirewallControlPolicy Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}
	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	d.Set("acl_uuid", parts[0])
	d.Set("direction", parts[1])
	d.Set("order", formatInt(object["Order"]))

	return nil
}

func resourceAlibabacloudStackCloudFirewallControlPolicyOrderDelete(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[DEBUG] Resource alibabacloudstack_cloud_firewall_control_policy_order [%s]  will not be deleted", d.Id())
	return nil
}
