package main
// learn encoding/json


import "encoding/json"
import "fmt"

//import "log"

//import "os"

func main() {
	boolT, _ := json.Marshal(true)
	fmt.Println(string(boolT))
	fmt.Printf("%s\n", boolT)

	intT, _ := json.Marshal(123)
	fmt.Println(string(intT))

	floatT, _ := json.Marshal(1.23)
	fmt.Println(string(floatT))
	fmt.Printf("%T\n", floatT)

	stringT, _ := json.Marshal("fastweb")
	fmt.Println(string(stringT))

	slc := []string{"apple", "peach", "pear"}
	sliceT, _ := json.Marshal(slc)
	fmt.Println(string(sliceT))

	mapj := map[string]int{"xiongwei": 90, "yaowei": 60, "tunwei": 95}
	mapT, _ := json.Marshal(mapj)
	fmt.Println(string(mapT))
	fmt.Printf("%T\n", mapT)

	type structA struct {
		Shengao int            `json:"shengao"`
		Shencai map[string]int `json:"shencai"`
	}

	girl := &structA{
		Shengao: 165,
		//	shencai: mapj,
		Shencai: map[string]int{"xiongwei": 90, "yaowei": 60, "tunwei": 95},
	}

	fmt.Println(girl)

	secret, err := json.Marshal(girl)
	if err != nil {
		//log.Fatalf("JSON marshaling failed:$s", err)
		fmt.Println("JSON marshaling failed!")
	}
	fmt.Println(secret)
	fmt.Println(string(secret))

	type Response struct {
		Page   int      `json:"page"`
		Fruits []string `json:"fruits"`
	}

	resd := &Response{
		Page:   1,
		Fruits: []string{"apple", "peach", "pear"},
	}
	fmt.Println(resd)
	resB, _ := json.Marshal(resd)
	fmt.Println(string(resB))

	byt := []byte(`{"num":6.13, "strs": ["a","b"]}`)

	var dat map[string]interface{}

	if err := json.Unmarshal(byt, &dat); err != nil {
		panic(err)

	}
	fmt.Println(dat)
	fmt.Printf("%T\n", dat["strs"])

	num := dat["num"].(float64)
	fmt.Println(num)

	strs := dat["strs"].([]interface{})
	strs1 := strs[0].(string)
	fmt.Println(strs1)

}
