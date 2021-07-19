package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"net/http"
)

type Content struct {
	Output string      `json:"output"`
	Input  string      `json:"input"`
	Layers []Component `json:"layers"`
}

type Config struct {
	Optimizer    string   `json:"optimizer"`
	LearningRate float64  `json:"learning_rate"`
	Loss         string   `json:"loss"`
	Metrics      []string `json:"metrics"`
	BatchSize    int      `json:"batch_size"`
	Epochs       int      `json:"epochs"`
	Output       string   `json:"output"`
}

type Component struct {
	Category string            `json:"category"`
	Type     string            `json:"type"`
	Name     string            `json:"name"`
	Input    *string           `json:"input"`
	Config   map[string]string `json:"config"`
}

const ImportTF = "import tensorflow as tf\n\n"
const TF = "tf"
const Keras = ".keras"
const Layer = ".layers"
const Math = ".math"

var category = map[string]string{
	"layer": TF + Keras + Layer,
	"math":  TF + Math,
}

func digitCheck(target string) bool {
	for i := 0; i < 10; i++ {
		if strings.Contains(target, string(i+'0')) {
			return true
		}
	}
	return false
}

// Generate Layer codes from content.json
func GenLayers(content Content) []string {
	var codes []string

	//contents, err := os.Open("./content.json")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//defer contents.Close()
	//
	//byteValue, _ := ioutil.ReadAll(contents)
	//
	//err = json.Unmarshal(byteValue, &content)
	//if err != nil {
	//	fmt.Println(err)
	//}

	// code converting
	for _, d := range content.Layers {
		layer := d.Name
		layer += " = "
		layer += category[d.Category] + "."
		layer += d.Type
		layer += "("

		i := 1
		for conf := range d.Config {
			var param string

			// if data is array like.
			if strings.Contains(d.Config[conf], ",") {
				param = fmt.Sprintf("%s=(%s)", conf, d.Config[conf])
			} else {
				if digitCheck(d.Config[conf]) {
					param = fmt.Sprintf("%s=%s", conf, d.Config[conf])
				} else {
					param = fmt.Sprintf("%s=\"%s\"", conf, d.Config[conf])
				}
			}
			layer += param
			if i < len(d.Config) {
				layer += ", "
			}
			i++
		}
		layer += ")"
		if d.Input != nil {
			layer += fmt.Sprintf("(%s)\n", *d.Input)
		} else {
			layer += "\n"
		}

		codes = append(codes, layer)
	}

	// create model.
	model := fmt.Sprintf("model = %s.Model(inputs=%s, outputs=%s)\n\n", TF+Keras, content.Input, content.Output)
	codes = append(codes, model)

	return codes
}

// generate compile codes from config.json
func GenConfig(config Config) []string {
	var codes []string

	//// model compile
	//configs, err := os.Open("./config.json")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//defer configs.Close()
	//
	//configByte,_ := ioutil.ReadAll(configs)
	//err = json.Unmarshal(configByte, &config)
	//if err != nil {
	//	fmt.Println(err)
	//}

	// get optimizer
	optimizer := fmt.Sprintf("%s.optimizers.%s(lr=%f)", TF+Keras, config.Optimizer, config.LearningRate)

	// get metrics
	var metrics string
	for i := 1; i <= len(config.Metrics); i++ {
		metrics += fmt.Sprintf("\"%s\"", config.Metrics[i-1])
		if i < len(config.Metrics) {
			metrics += ", "
		}
	}

	// get compile
	compile := fmt.Sprintf("model.compile(optimizer=%s, loss=\"%s\", metrics=[%s])\n", optimizer, config.Loss, metrics)
	codes = append(codes, compile)

	return codes
}

func GenerateCode(content Content, config Config) {
	var codes []string
	codes = append(codes, ImportTF)
	codes = append(codes, GenLayers(content)...)
	codes = append(codes, GenConfig(config)...)

	// create python file
	py, err := os.Create("test.py")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer py.Close()

	fileSize := 0
	for _, code := range codes {
		n, err := py.Write([]byte(code))
		if err != nil {
			fmt.Println(err)
			return
		}
		fileSize += n
	}
	fmt.Printf("Code converting is finish with %d bytes size", fileSize)
}

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.POST("/generate", func(c echo.Context) error {
		var content Content
		var config Config

		// binding JSON
		err := c.Bind(content)
		if err != nil {
			fmt.Println(err)
		}

		err = c.Bind(config)
		if err != nil {
			fmt.Println(err)
		}

		// return fIle.
		return c.Attachment("./test.py", "model.py")
	})

	e.Logger.Fatal(e.Start(":80"))
}