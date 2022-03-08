package nyahttphandle

import (
	"fmt"
	"net/http"

	"github.com/tidwall/gjson"
)

// type NyahttpHandle NyahttpHandleT
// type NyahttpHandleT struct {
// 	err error
// }

func Init(confstr string, handlers ...func(http.ResponseWriter, *http.Request)) error {
	listenandserve := gjson.Get(confstr, "listenandserve")
	if !listenandserve.Exists() {
		return fmt.Errorf("no config :listenandserve")
	}
	if listenandserve.Type.String() != "String" {
		return fmt.Errorf("config 'listenandserve' is not string")
	}
	suburl := gjson.Get(confstr, "suburl")
	if !suburl.Exists() {
		return fmt.Errorf("no config :suburl")
	}
	if suburl.Type.String() != "String" {
		return fmt.Errorf("config 'suburl' is not string")
	}
	patterns := gjson.Get(confstr, "patterns")
	if !patterns.Exists() {
		return fmt.Errorf("no config :patterns")
	}
	if patterns.Type.String() != "JSON" {
		return fmt.Errorf("config 'patterns' is not JSON")
	}
	if len(patterns.Array()) != len(handlers) {
		return fmt.Errorf("the number of config 'patterns' and 'handlers' is not equal")
	}
	for i, v := range handlers {
		if patterns.Array()[i].Type.String() != "String" {
			fmt.Println(patterns.Array()[i], " is not string")
			continue
		}
		pattern := suburl.String() + patterns.Array()[i].String()
		http.HandleFunc(pattern, v)
	}
	//设置监听的端口
	err := http.ListenAndServe(listenandserve.String(), nil)
	if err != nil {
		return err
	}
	return nil
}
