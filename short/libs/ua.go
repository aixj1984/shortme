package libs

import (
	"regexp"
	"strings"
)

//return source model
func GetDeviceInfoFromUa(ua string) (string, string, error) {
	//str := "Mozilla/5.0 (Linux; U; Android 8.1.0; zh-cn; AWM-A0 Build/G66T1901280CN00MP4) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/53.0.2785.146 Mobile Safari/537.36 XiaoMi/MiuiBrowser/9.4.12"
	model := ""
	version := ""
	reg, err := regexp.Compile(`\(.+?\)`)
	if err != nil {
		return model, version, err
	}
	remark := reg.FindString(ua)

	reg, err = regexp.Compile(`;\s?([^;]+?)\sBuild\/(.*)[;)]`)
	if err != nil {
		if strings.Contains(ua, "SKW-A0") {
			model = "SKW-A0"
		} else if strings.Contains(ua, "SKR-A0") {
			model = "SKR-A0"
		} else if strings.Contains(ua, "AWM-A0") {
			model = "AWM-A0"
		}

		if len(model) > 0 {
			return model, version, nil
		} else {
			return model, version, err
		}

	}
	modelstr := reg.FindStringSubmatch(remark)
	if len(modelstr) == 3 {
		model = modelstr[1]

		versions := strings.Split(modelstr[2], ";")

		version = versions[0]
	}

	return model, version, nil
}
