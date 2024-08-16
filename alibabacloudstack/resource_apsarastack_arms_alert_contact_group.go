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
)

func resourceAlibabacloudStackArmsAlertContactGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlibabacloudStackArmsAlertContactGroupCreate,
		Read:   resourceAlibabacloudStackArmsAlertContactGroupRead,
		Update: resourceAlibabacloudStackArmsAlertContactGroupUpdate,
		Delete: resourceAlibabacloudStackArmsAlertContactGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"alert_contact_group_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"contact_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceAlibabacloudStackArmsAlertContactGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AlibabacloudStackClient)
	var response map[string]interface{}
	action := "CreateAlertContactGroup"
	request := make(map[string]interface{})
	conn, err := client.NewArmsClient()
	if err != nil {
		return WrapError(err)
	}
	request["ContactGroupName"] = d.Get("alert_contact_group_name")
	if v, ok := d.GetOk("contact_ids"); ok {
		request["ContactIds"] = convertArrayToString(v.(*schema.Set).List(), " ")
	}
	request["RegionId"] = client.RegionId
	request["Product"] = "ARMS"
	request["OrganizationId"] = client.Department
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err = conn.DoRequest(StringPointer(action), nil, StringPointer("POST"), StringPointer("2019-08-08"), StringPointer("AK"), nil, request, &util.RuntimeOptions{IgnoreSSL: tea.Bool(client.Config.Insecure)})
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
		return WrapErrorf(err, DefaultErrorMsg, "alibabacloudstack_arms_alert_contact_group", action, AlibabacloudStackSdkGoERROR)
	}

	d.SetId(fmt.Sprint(response["ContactGroupId"]))

	return resourceAlibabacloudStackArmsAlertContactGroupRead(d, meta)
}
func resourceAlibabacloudStackArmsAlertContactGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AlibabacloudStackClient)
	armsService := ArmsService{client}
	object, err := armsService.DescribeArmsAlertContactGroup(d.Id())
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[DEBUG] Resource alibabacloudstack_arms_alert_contact_group armsService.DescribeArmsAlertContactGroup Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}
	d.Set("alert_contact_group_name", object["ContactGroupName"])
	contactIdsItems := make([]string, 0)
	if contacts, ok := object["Contacts"]; ok && contacts != nil {
		for _, contactsItem := range contacts.([]interface{}) {
			if contactId, ok := contactsItem.(map[string]interface{})["ContactId"]; ok && contactId != nil {
				contactIdsItems = append(contactIdsItems, fmt.Sprint(contactId))
			}
		}
	}
	d.Set("contact_ids", contactIdsItems)
	return nil
}
func resourceAlibabacloudStackArmsAlertContactGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AlibabacloudStackClient)
	var response map[string]interface{}
	update := false
	request := map[string]interface{}{
		"ContactGroupId": d.Id(),
	}
	if d.HasChange("alert_contact_group_name") {
		update = true
	}
	request["ContactGroupName"] = d.Get("alert_contact_group_name")
	request["RegionId"] = client.RegionId
	request["Product"] = "ARMS"
	request["OrganizationId"] = client.Department
	if d.HasChange("contact_ids") {
		update = true
	}
	if v, ok := d.GetOk("contact_ids"); ok {
		request["ContactIds"] = convertArrayToString(v.(*schema.Set).List(), " ")
	}
	if update {
		action := "UpdateAlertContactGroup"
		conn, err := client.NewArmsClient()
		if err != nil {
			return WrapError(err)
		}
		wait := incrementalWait(3*time.Second, 3*time.Second)
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			response, err = conn.DoRequest(StringPointer(action), nil, StringPointer("POST"), StringPointer("2019-08-08"), StringPointer("AK"), nil, request, &util.RuntimeOptions{IgnoreSSL: tea.Bool(client.Config.Insecure)})
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
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabacloudStackSdkGoERROR)
		}
	}
	return resourceAlibabacloudStackArmsAlertContactGroupRead(d, meta)
}
func resourceAlibabacloudStackArmsAlertContactGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AlibabacloudStackClient)
	action := "DeleteAlertContactGroup"
	var response map[string]interface{}
	conn, err := client.NewArmsClient()
	if err != nil {
		return WrapError(err)
	}
	request := map[string]interface{}{
		"ContactGroupId": d.Id(),
	}

	request["RegionId"] = client.RegionId
	request["Product"] = "ARMS"
	request["OrganizationId"] = client.Department
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		response, err = conn.DoRequest(StringPointer(action), nil, StringPointer("POST"), StringPointer("2019-08-08"), StringPointer("AK"), nil, request, &util.RuntimeOptions{IgnoreSSL: tea.Bool(client.Config.Insecure)})
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
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabacloudStackSdkGoERROR)
	}
	return nil
}
