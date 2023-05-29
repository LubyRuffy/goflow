package gocodefuncs

import (
	"github.com/LubyRuffy/goflow/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestToExcel(t *testing.T) {
	// 读excel-》json
	filename := writeSampleExcelFile()
	fr := ExcelToJson(&testRunner{
		T:        t,
		lastFile: filename,
	}, map[string]interface{}{})

	// json=》写excel
	result := ToExcel(&testRunner{
		T:        t,
		lastFile: fr.OutFile,
	}, map[string]interface{}{
		"rawFormat": true,
	})

	// 再读
	fr = ExcelToJson(&testRunner{
		T:        t,
		lastFile: result.Artifacts[0].FilePath,
	}, map[string]interface{}{})
	f, err := utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, string(f), `{"Sheet1":[["IP","域名"],["1.1.1.1","a.com"]],"Sheet2":[null,["Hello world."]]}`)

	// json=》写excel
	json := `{"Sheet1":[["IP"],["1.1.1.1"],["2.2.2.2"]]}`
	fn, err := utils.WriteTempFile(".json", func(f *os.File) error {
		_, err := f.WriteString(json)
		return err
	})
	result = ToExcel(&testRunner{
		T:        t,
		lastFile: fn,
	}, map[string]interface{}{
		"rawFormat": true,
	})
	// 再读
	fr = ExcelToJson(&testRunner{
		T:        t,
		lastFile: result.Artifacts[0].FilePath,
	}, map[string]interface{}{})
	f, err = utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, string(f), `{"Sheet1":[["IP"],["1.1.1.1"],["2.2.2.2"]]}`)

	// 写合并的表格
	json = `{"Sheet1":[["t1","t2","t3"],["dct11","dct12","dct13"],["dct11","dct22","dct13"]],"Sheet2":[["t1","t2","t3"],["dct11","dct11","dct12"],["dct11","dct11","dct22"]],"_merged_Sheet1":[["A2:A3","dct11"],["C2:C3","dct13"]],"_merged_Sheet2":[["A2:B3","dct11"]]}`
	fn, err = utils.WriteTempFile(".json", func(f *os.File) error {
		_, err := f.WriteString(json)
		return err
	})
	result = ToExcel(&testRunner{
		T:        t,
		lastFile: fn,
	}, map[string]interface{}{
		"rawFormat": true,
	})
	// 再读
	fr = ExcelToJson(&testRunner{
		T:        t,
		lastFile: result.Artifacts[0].FilePath,
	}, map[string]interface{}{})
	f, err = utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, string(f), `{"Sheet1":[["t1","t2","t3"],["dct11","dct12","dct13"],["dct11","dct22","dct13"]],"Sheet2":[["t1","t2","t3"],["dct11","dct11","dct12"],["dct11","dct11","dct22"]],"_merged_Sheet1":[["A2:A3","dct11"],["C2:C3","dct13"]],"_merged_Sheet2":[["A2:B3","dct11"]]}`)
}

func TestToExcel1(t *testing.T) {
	type args struct {
		params    map[string]interface{}
		inputJson string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "测试raw格式输入",
			args: args{
				params: map[string]interface{}{
					"rawFormat": true,
					"insertPic": false,
				},
				inputJson: "{\"Sheet1\":[[\"IP\",\"47.106.125.91\",\"47.106.125.91\",\"47.106.125.91\",\"47.106.125.91\"],[\"地理位置\",\"广东省广州市\",\"注册机构\",\"阿里云\",\"阿里云\"],[\"网络类型\",\"数据中心\",\"经纬度\",\"113.307650,23.120049\",\"113.307650,23.120049\"],[\"ASN\",\"AS37963CNNIC-ALIBABA-CN-NET-AP\",\"IDC服务器\",\"是\",\"是\"],[\"网络资产\",\"22\",\"TCP/SSH\",\"OpenSSH\",\"SSH-2.0-OpenSSH_7.4\"],[\"网络资产\",\"443\",\"TCP/HTTP\",\"NGINX,digicert-Cert\",\"HTTP/1.1502BadGateway\\nConnection:close\"],[\"网络资产\",\"80\",\"TCP/HTTP\",\"NGINX\",\"HTTP/1.1200OK\\nConnection:close\"],[\"网络资产\",\"9200\",\"TCP/ELASTIC\",\"Log4j2,Elasticsearch\",\"HTTP/1.0200OK\\nX-elastic-product:Elasticsearch\"],[\"网络资产\",\"11001\",\"TCP/HTTP\",\"\",\"HTTP/1.1400BadRequest\\nContent-Type:text/html;charset=us-ascii\"],[\"网络资产\",\"3000\",\"TCP/HTTP\",\"\",\"HTTP/1.1200OK\\nAccess-Control-Allow-Origin:*\"],[\"网络资产\",\"6001\",\"TCP/UNKNOW\",\"\",\"\\\\x00\\\\x00\\\\x00\\\\x01\\\\x00\"],[\"网络资产\",\"9003\",\"TCP/HTTP\",\"\",\"HTTP/1.1200OK\\nConnection:close\"],[\"根域资产（半年内）\",\"enyamusic.cn\",\"\",\"postmaster@enyamusic.pro\",\"广州恩雅创新科技有限公司\"],[\"根域资产（半年内）\",\"makedingge.com\",\"+86.95187\",\"DomainAbuse@service.aliyun.com\"],[\"IP情报\",\"02-17-17\",\"06-01-18\",\"垃圾邮件\",\"过期\"],[\"Web页面\",\"443\",\"资产测绘标题（502BadGateway）\",\"目前标题（502BadGateway）\"],[\"Web页面\",\"80\",\"资产测绘标题（Welcometonginx!）\",\"目前标题（Welcometonginx!）\"],[\"Web页面\",\"11001\",\"资产测绘标题（无）\",\"目前标题（无）\",\"无法截图\"],[\"Web页面\",\"3000\",\"资产测绘标题（无）\",\"目前标题（无）\",\"无法截图\"],[\"Web页面\",\"9003\",\"资产测绘标题（无）\",\"目前标题（无）\",\"无法截图\"],[\"证书\",\"443\",\"\",\"*.enyamusic.cn\",\"*.enyamusic.cn\"],[\"C段Web资产标题排序\",\"口袋客app（771）\",\"口袋客app（771）\",\"口袋客app（771）\",\"口袋客app（771）\"],[\"C段Web资产标题排序\",\"301MovedPermanently（50）\",\"301MovedPermanently（50）\",\"301MovedPermanently（50）\",\"301MovedPermanently（50）\"],[\"C段Web资产标题排序\",\"403Forbidden（49）\",\"403Forbidden（49）\",\"403Forbidden（49）\",\"403Forbidden（49）\"],[\"C段Web资产标题排序\",\"安全入口校验失败（26）\",\"安全入口校验失败（26）\",\"安全入口校验失败（26）\",\"安全入口校验失败（26）\"],[\"C段Web资产标题排序\",\"NotFound（21）\",\"NotFound（21）\",\"NotFound（21）\",\"NotFound（21）\"],[\"IP和域名网络情报\",\"enyamusic.cn\",\"sadf\",\"sadf\",\"sadf\"],[\"IP和域名网络情报\",\"makedingge.com\",\"asadf\",\"asadf\",\"asadf\"],[\"IP和域名网络情报\",\"103.117.102.89\",\"fdfasf\",\"fdfasf\",\"fdfasf\"]],\"_merged_Sheet1\":[[\"A27:A29\",\"IP和域名网络情报\"],[\"C27:E27\",\"sadf\"],[\"C28:E28\",\"asadf\"],[\"C29:E29\",\"fdfasf\"],[\"A22:A26\",\"C段Web资产标题排序\"],[\"A13:A14\",\"根域资产（半年内）\"],[\"D21:E21\",\"*.enyamusic.cn\"],[\"B26:E26\",\"NotFound（21）\"],[\"A16:A20\",\"Web页面\"],[\"B25:E25\",\"安全入口校验失败（26）\"],[\"B24:E24\",\"403Forbidden（49）\"],[\"B23:E23\",\"301MovedPermanently（50）\"],[\"B22:E22\",\"口袋客app（771）\"],[\"D3:E3\",\"113.307650,23.120049\"],[\"B1:E1\",\"47.106.125.91\"],[\"D4:E4\",\"是\"],[\"D2:E2\",\"阿里云\"],[\"A5:A12\",\"网络资产\"]]}",
			},
		},
		{
			name: "测试自动合并输入",
			args: args{
				params: map[string]interface{}{
					"rawFormat": false,
					"insertPic": false,
					"autoMerge": true,
				},
				inputJson: "{\"Sheet1\":[[\"IP\",\"47.106.125.91\",\"47.106.125.91\",\"47.106.125.91\",\"47.106.125.91\"],[\"地理位置\",\"广东省广州市\",\"注册机构\",\"阿里云\",\"阿里云\"],[\"网络类型\",\"数据中心\",\"经纬度\",\"113.307650,23.120049\",\"113.307650,23.120049\"],[\"ASN\",\"AS37963CNNIC-ALIBABA-CN-NET-AP\",\"IDC服务器\",\"是\",\"是\"],[\"网络资产\",\"22\",\"TCP/SSH\",\"OpenSSH\",\"SSH-2.0-OpenSSH_7.4\"],[\"网络资产\",\"443\",\"TCP/HTTP\",\"NGINX,digicert-Cert\",\"HTTP/1.1502BadGateway\\nConnection:close\"],[\"网络资产\",\"80\",\"TCP/HTTP\",\"NGINX\",\"HTTP/1.1200OK\\nConnection:close\"],[\"网络资产\",\"9200\",\"TCP/ELASTIC\",\"Log4j2,Elasticsearch\",\"HTTP/1.0200OK\\nX-elastic-product:Elasticsearch\"],[\"网络资产\",\"11001\",\"TCP/HTTP\",\"\",\"HTTP/1.1400BadRequest\\nContent-Type:text/html;charset=us-ascii\"],[\"网络资产\",\"3000\",\"TCP/HTTP\",\"\",\"HTTP/1.1200OK\\nAccess-Control-Allow-Origin:*\"],[\"网络资产\",\"6001\",\"TCP/UNKNOW\",\"\",\"\\\\x00\\\\x00\\\\x00\\\\x01\\\\x00\"],[\"网络资产\",\"9003\",\"TCP/HTTP\",\"\",\"HTTP/1.1200OK\\nConnection:close\"],[\"根域资产（半年内）\",\"enyamusic.cn\",\"\",\"postmaster@enyamusic.pro\",\"广州恩雅创新科技有限公司\"],[\"根域资产（半年内）\",\"makedingge.com\",\"+86.95187\",\"DomainAbuse@service.aliyun.com\"],[\"IP情报\",\"02-17-17\",\"06-01-18\",\"垃圾邮件\",\"过期\"],[\"Web页面\",\"443\",\"资产测绘标题（502BadGateway）\",\"目前标题（502BadGateway）\"],[\"Web页面\",\"80\",\"资产测绘标题（Welcometonginx!）\",\"目前标题（Welcometonginx!）\"],[\"Web页面\",\"11001\",\"资产测绘标题（无）\",\"目前标题（无）\",\"无法截图\"],[\"Web页面\",\"3000\",\"资产测绘标题（无）\",\"目前标题（无）\",\"无法截图\"],[\"Web页面\",\"9003\",\"资产测绘标题（无）\",\"目前标题（无）\",\"无法截图\"],[\"证书\",\"443\",\"\",\"*.enyamusic.cn\",\"*.enyamusic.cn\"],[\"C段Web资产标题排序\",\"口袋客app（771）\",\"口袋客app（771）\",\"口袋客app（771）\",\"口袋客app（771）\"],[\"C段Web资产标题排序\",\"301MovedPermanently（50）\",\"301MovedPermanently（50）\",\"301MovedPermanently（50）\",\"301MovedPermanently（50）\"],[\"C段Web资产标题排序\",\"403Forbidden（49）\",\"403Forbidden（49）\",\"403Forbidden（49）\",\"403Forbidden（49）\"],[\"C段Web资产标题排序\",\"安全入口校验失败（26）\",\"安全入口校验失败（26）\",\"安全入口校验失败（26）\",\"安全入口校验失败（26）\"],[\"C段Web资产标题排序\",\"NotFound（21）\",\"NotFound（21）\",\"NotFound（21）\",\"NotFound（21）\"],[\"IP和域名网络情报\",\"enyamusic.cn\",\"sadf\",\"sadf\",\"sadf\"],[\"IP和域名网络情报\",\"makedingge.com\",\"asadf\",\"asadf\",\"asadf\"],[\"IP和域名网络情报\",\"103.117.102.89\",\"fdfasf\",\"fdfasf\",\"fdfasf\"]]}",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestRunner(t, tt.args.inputJson)
			got := ToExcel(p, tt.args.params)
			assert.NotNil(t, got.Artifacts)
		})
	}
}
