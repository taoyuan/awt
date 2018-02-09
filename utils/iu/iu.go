package iu

import (
	"go-awt/infra/osencap"
	"encoding/json"
	"github.com/thoas/go-funk"
	"errors"
	"strconv"
	"strings"
)

type Hardware struct {
	Id            string                 `json:"id"`
	Class         string                 `json:"class"`
	Claimed       bool                   `json:"claimed"`
	Description   string                 `json:"description"`
	Product       string                 `json:"product"`
	Serial        string                 `json:"serial"`
	Width         int                    `json:"width"`
	Physid        int                    `json:"physid"`
	Logicalname   string                 `json:"logicalname"`
	Businfo       string                 `json:"businfo"`
	Units         string                 `json:"units"`
	Size          int                    `json:"size"`
	Capacity      int                    `json:"capacity"`
	Configuration map[string]interface{} `json:"configuration"`
	Capabilities  map[string]interface{} `json:"capabilities"`
	Children      []Hardware             `json:"children"`
	Tags		  []string
}

func Lshw(class string) (*Hardware, error) {
	args := []string{"-json"}
	if class != "" {
		args = append(args, "-class", class)
	}
	out, err := osencap.Exec("lshw", args...)
	if err != nil {
		return nil, err
	}

	if class != "" {
		out = `{"children":[` + out + `]}`
	}
	var hw Hardware
	json.Unmarshal([]byte(out), &hw)
	return &hw, nil
}

func LsWlans() ([]Hardware, error)  {
	root, err := Lshw("network")
	if err != nil {
		return nil, err
	}
	var seqOnboard = 0
	var answer []Hardware
	for _, hw := range root.Children {
		if hw.Capabilities["wireless"] == "Wireless-LAN" {
			if hw.Logicalname != "" {
				hw.Tags = append(hw.Tags, hw.Logicalname)
			} else {
				continue
			}

			if hw.Businfo == "" {
				hw.Tags = append(hw.Tags, "default")
				hw.Tags = append(hw.Tags, "onboard")
				hw.Tags = append(hw.Tags, "onboard" + strconv.Itoa(seqOnboard))
				seqOnboard = seqOnboard + 1
			} else {
				hw.Tags = append(hw.Tags, hw.Businfo)
			}

			answer = append(answer, hw)
		}
	}
	return answer, nil
}

func ResolveIface(iface string) (string, error) {
	wlans, err := LsWlans()
	if err != nil {
		return "", err
	}

	var ifaceToUse = iface
	if ifaceToUse == "" {
		ifaceToUse = "default"
	}

	for _, hw := range wlans {
		if funk.Contains(hw.Tags, ifaceToUse) {
			return hw.Logicalname, nil
		}
	}

	for _, hw := range wlans {
		if strings.HasPrefix(hw.Businfo, ifaceToUse) {
			return hw.Logicalname, nil
		}
	}

	if iface == "" && len(wlans) > 0 {
		return wlans[0].Logicalname, nil
	}

	return "", errors.New("iface \"" + iface + "\" can not be found")
}
