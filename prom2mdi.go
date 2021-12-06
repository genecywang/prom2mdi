package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	MDI_URL string `mapstructure:"mdi_url"`
	Token   string `mapstructure:"token"`
	Targets string `mapstructure:"targets"`
}

type Message struct {
	Token   string `json:"token"`
	Targets string `json:"targets"`
	Content string `json:"content"`
}

type Alert struct {
	Status      string            `jons:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:annotations`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      time.Time         `json:"endsAt"`
}

type Notification struct {
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	Status            string            `json:"status"`
	Receiver          string            `json:receiver`
	GroupLabels       map[string]string `json:groupLabels`
	CommonLabels      map[string]string `json:commonLabels`
	CommonAnnotations map[string]string `json:commonAnnotations`
	ExternalURL       string            `json:externalURL`
	Alerts            []Alert           `json:alerts`
}

var Conf = new(Config)

func sendAlertToHooks(url string, msg string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("err")
		}
	}()

	req, err := http.NewRequest("POST", url, strings.NewReader(msg))
	if err != nil {
		fmt.Println(err)
	}
	client := &http.Client{}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Response Body:", string(body))

}

func mdiJSONPayload(token string, targets string, content string) string {
	var message Message

	message.Targets = targets
	message.Token = token
	message.Content = content

	jsons, err := json.Marshal(message)
	if err != nil {
		fmt.Println(err)
	}
	mesg := string(jsons)
	return mesg
}

func readTmplToString(filename string, data interface{}) (string, error) {

	t := template.Must(
		template.New(filename).Funcs(sprig.TxtFuncMap()).ParseGlob("*.tmpl"))

	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	t.Parse(
		string(content),
	)

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		return "", err
	}

	return tpl.String(), nil
}

func bindAlertManagerPost(c *gin.Context) {
	var notification Notification

	err := c.BindJSON(&notification)
	fmt.Printf("%#v\n", notification)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else {
		fmt.Printf(">>> [%s] Bind OK.\n", c.ClientIP())

	}

	alertMesg, err := readTmplToString(viper.GetString("template"), notification)
	if err != nil {
		fmt.Println(err)
	}

	payload := mdiJSONPayload(
		Conf.Token,
		Conf.Targets,
		alertMesg)

	fmt.Println(payload)
	sendAlertToHooks(Conf.MDI_URL, payload)
}

func main() {
	pflag.String("port", "9000", "Port for listening to static web page.")
	pflag.String("name", "Alert", "API Hook page. ex: http://${URL}/hooks/${name}")
	pflag.String("config", "configs/config.yaml", "Config file path.")
	pflag.String("template", "configs/mdi.tmpl", "Template file path.")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetConfigFile(viper.GetString("config"))
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	if err := viper.Unmarshal(Conf); err != nil {
		panic(fmt.Errorf("unmarshal conf failed. err: %s", err))
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("The Config file has been modified.", e.Name)
		if err := viper.Unmarshal(Conf); err != nil {
			panic(fmt.Errorf("unmarshal conf failed. err: %s", err))
		}
	})

	// Parse Template file
	_, t_err := template.New(viper.GetString("template")).
		Funcs(sprig.TxtFuncMap()).
		ParseFiles(viper.GetString("template"))

	if t_err != nil {
		panic(t_err)
	}

	//----
	t := gin.Default()
	t.POST("/hooks/"+viper.GetString("name"), bindAlertManagerPost)
	t.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Status OK")
	})
	t.Run(":" + viper.GetString("port"))
	if err != nil {
		panic(fmt.Errorf("cannot listen on %s port, err:%s",
			viper.GetString("port"),
			err))
	}
	//---

}
