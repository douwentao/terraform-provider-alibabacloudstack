package alibabacloudstack

import (
	"strings"

	"github.com/PaesslerAG/jsonpath"
	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/aliyun/terraform-provider-alibabacloudstack/alibabacloudstack/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAlibabacloudStackNasProtocols() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlibabacloudStackNasProtocolsRead,

		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Capacity",
					"Performance",
				}, false),
			},
			"zone_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"protocols": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAlibabacloudStackNasProtocolsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AlibabacloudStackClient)
	action := "DescribeZones"
	var response map[string]interface{}
	request := make(map[string]interface{})
	request["RegionId"] = client.RegionId
	request["Product"] = "Nas"
	request["OrganizationId"] = client.Department
	conn, err := client.NewNasClient()
	if err != nil {
		return WrapError(err)
	}

	runtime := util.RuntimeOptions{IgnoreSSL: tea.Bool(client.Config.Insecure)}
	runtime.SetAutoretry(true)
	response, err = conn.DoRequest(StringPointer(action), nil, StringPointer("POST"), StringPointer("2017-06-26"), StringPointer("AK"), nil, request, &runtime)
	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alibabacloudstack_nas_protocols", action, AlibabacloudStackSdkGoERROR)
	}
	addDebug(action, response, request)
	resp, err := jsonpath.Get("$.Zones.Zone", response)
	if err != nil {
		return WrapErrorf(err, FailedGetAttributeMsg, action, "$.Zones.Zone", response)
	}
	var nasProtocol [][]interface{}
	result, _ := resp.([]interface{})
	for _, v := range result {
		item := v.(map[string]interface{})
		if v, ok := d.GetOk("zone_id"); ok && v.(string) != "" && item["ZoneId"].(string) != v.(string) {
			continue
		}
		if v, ok := d.GetOk("type"); ok {
			var strVal = Trim(v.(string))
			if strVal == "Performance" || strVal == "Capacity" {
				var clusterVal = item["Clusters"].(map[string]interface{})["Cluster"].([]interface{})
				var instanceTypes = clusterVal[0].(map[string]interface{})["InstanceTypes"].(map[string]interface{})["InstanceType"].([]interface{})
				var newProtocol []interface{}
				for _, b := range instanceTypes {
					a := b.(map[string]interface{})["ProtocolType"]
					newProtocol = append(newProtocol, a)
				}
				if len(newProtocol) == 0 {
					continue
				} else {
					nasProtocol = append(nasProtocol, newProtocol)
				}
			}
		}
	}
	var s []string
	var ids []string
	for _, val := range nasProtocol {
		for _, protocol := range val {
			s = append(s, strings.ToUpper(protocol.(string)))
			ids = append(ids, protocol.(string))
		}
	}
	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("protocols", s); err != nil {
		return WrapError(err)
	}
	// create a json file in current directory and write data source to it.
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}
	return nil
}
