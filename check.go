package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var i = 0

const token string = "<input discord bot token>"

var BotID string

type bod struct {
	Integrations []triggers `json:"integrations"`
	Description string `json:"description"`
}

type triggers struct {
	Triggers      []trigpart `json:"triggers"`
	CookieDomain  string     `json:"cookieDomain"`
	RedirectLogic string     `json:"redirectLogic"`
}

type trigpart struct {
	TriggerParts    []parts `json:"triggerParts"`
	LogicalOperator string  `json:"logicalOperator"`
}

type parts struct {
	Value         string `json:"valueToCompare"`
	ValuesToCompare []string `json:"valuesToCompare"`
	Operator      string `json:"operator"`
	UrlPart       string `json:"urlPart"`
	ValidatorType string `json:"validatorType"`
	IsNegative    bool   `json:"isNegative"`
	IsIgnoreCase  bool   `json:"isIgnoreCase"`
}

var proxylist = createsplice("proxies.txt")
var proxyiteration = 0

func rotateproxy(list []string) (string, string, string) {
	if proxyiteration == len(list)-1 {
		proxyiteration = 0
	}
	proxy := list[proxyiteration]
	splice := strings.Split(proxy, ":")
	temp := strings.Fields(splice[3])
	proxyiteration++
	return splice[0] + ":" + splice[1], splice[2], temp[0]
}

func createsplice(file string) []string {
	txt, err := os.Open(file)
	txt.Sync()
	if err != nil {
		log.Fatal(err)
	}
	byte, err := ioutil.ReadAll(txt)
	if err != nil {
		log.Fatal(err)
	}

	return strings.Split(string(byte), "\n")
}

func writedic(word string, file string) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if _, err = f.WriteString(word + "\n"); err != nil {
		log.Fatal(err)
	}
}

func rewrite(dic string) {

	output := []byte(dic)

	if err := ioutil.WriteFile("dictionary.txt", output, 0666); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func checkdeleted(current []string, dictionary []string) ([]string, string) {
	for j := range dictionary {
		for k := range current {
			if strings.Compare(current[k], dictionary[j]) == 0 { //current sku still exists
				break
			} else if k == len(current)-1 { //dictionary word does not exist in current queueit
				//rewrite file with new words using current splice
				for v, str := range dictionary {
					if str == dictionary[j] { //dictionary j is not in current splice
						dictionary = append(dictionary[:v], dictionary[v+1:]...)
						var text string
						for l := range dictionary {
							text = text + dictionary[l] + "\n"
						}
						//rewrite text into file
						rewrite(text)

						return dictionary, str

					}
				}
			}
		}
	}
	return nil, "" //do nothing if no deletion
}

func createdeletion(storeid string, description string) *discordgo.WebhookParams {
	if strings.Compare(storeid, "") == 0 {
		fmt.Println("empty string")
		return nil
	}
	tempstore := strings.Split(storeid, ":")
	sku := tempstore[0]
	store := tempstore[1]
	var fields []*discordgo.MessageEmbedField
	temp1 := &discordgo.MessageEmbedField{
		Name:  "SKU",
		Value: sku,
	}
	temp2 := &discordgo.MessageEmbedField{
		Name:  "Store",
		Value: store,
	}
	fields = append(fields, temp2)
	fields = append(fields, temp1)

	embed := &discordgo.MessageEmbed{
		Title:       ":rotating_light: Queue-it Has Been Removed From " + sku + " On " + store + " :rotating_light:",
		URL:         "https://www." + store + "/product/~/" + sku + ".html",
		Description: "No More Queue for " + sku + "!!!! \nDescription: " + "***"+description+"***",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://images.footlocker.com/is/image/EBFL2/" + sku + "_a1",
		},
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Property of Nuggie Labs",
			IconURL: "",
		},
		Color: 47349,
	}
	var id []*discordgo.MessageEmbed
	id = append(id, embed)
	webhook := discordgo.WebhookParams{
		Username: "Queue-it Monitor",
		Embeds:   id,
	}

	return &webhook
}

func checkarray(dg *discordgo.Session, body bod, k int, trig int, part int, current []string, dictionary []string) ([]string, []string){
	store := body.Integrations[k].CookieDomain
	for i := range body.Integrations[k].Triggers[trig].TriggerParts[part].ValuesToCompare{
		sku := body.Integrations[k].Triggers[trig].TriggerParts[part].ValuesToCompare[i]
		current = append(current, sku+":"+store)
		for x := range dictionary {
			tempdic := strings.Split(dictionary[x], ":")
			if strings.Compare(tempdic[0], sku) == 0 && len(tempdic) > 1 && strings.Compare(tempdic[1], store) == 0 {
				//ALREADY EXISTS
				break
			} else if x == len(dictionary) - 1 {
				//ADD TO DICTIONARY
				dictionary = append(dictionary, sku+":"+store)
				writedic(sku+":"+store, "dictionary.txt")
				fmt.Println(sku + ":" + store + " ADDED")

				domain := body.Integrations[k].CookieDomain

				webhook := createalert(sku, domain, body.Description)
				_, _ = dg.WebhookExecute("<discord webhook>", "<discord webhook>", true, webhook)

				advwebhook := createspecialalert(sku, body.Integrations[k].Triggers[trig].TriggerParts[part].ValuesToCompare[i], domain, body.Integrations[k].RedirectLogic, body.Integrations[k].Triggers[0].LogicalOperator, body.Integrations[k].Triggers[trig].TriggerParts[part].Operator, body.Integrations[k].Triggers[trig].TriggerParts[part].UrlPart, body.Integrations[k].Triggers[trig].TriggerParts[part].ValidatorType, body.Integrations[k].Triggers[trig].TriggerParts[part].IsNegative, body.Integrations[k].Triggers[trig].TriggerParts[part].IsIgnoreCase, body.Description)
				_, _ = dg.WebhookExecute("<discord webhook>", "<discord webhook>", true, advwebhook) 
				time.Sleep(time.Second)
			}
		}
		time.Sleep(time.Second)
	}

	return current, dictionary
}

func createalert(sku string, domain string, description string) *discordgo.WebhookParams {
	var fields []*discordgo.MessageEmbedField
	temp1 := &discordgo.MessageEmbedField{
		Name:  "Domain",
		Value: domain,
	}
	temp2 := &discordgo.MessageEmbedField{
		Name:  "SKU",
		Value: sku,
	}
	fields = append(fields, temp1)
	fields = append(fields, temp2)

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: domain,
			URL:  "https://" + domain + "/",
		},
		Title:       ":rotating_light: Queue-it Added To A New Product :rotating_light:",
		URL:         "https://" + domain + "/product/~/" + sku + ".html",
		Description: "A New Product has been added to Queue-it!!!! \nDescription: " + "***"+description+"***",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://images.footlocker.com/is/image/EBFL2/" + sku + "_a1",
		},
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Property of Nuggie Labs",
			IconURL: "",
		},
		Color: 14902784,
	}
	var id []*discordgo.MessageEmbed
	id = append(id, embed)
	webhook := discordgo.WebhookParams{
		Username: "Queue-it Monitor",
		Embeds:   id,
	}

	return &webhook
}

func createspecialalert(sku string, value string, domain string, redirectlogic string, logicaloperator string, operator string, urlpart string, validatortype string, isnegative bool, isignorecase bool, description string) *discordgo.WebhookParams {
	var fields []*discordgo.MessageEmbedField
	temp1 := &discordgo.MessageEmbedField{
		Name:   "Domain",
		Value:  domain,
		Inline: true,
	}
	temp2 := &discordgo.MessageEmbedField{
		Name:   "SKU",
		Value:  sku,
		Inline: true,
	}
	separator := &discordgo.MessageEmbedField{
		Name:  "~~~~~~~~~~~~~~~~~~~~~~~~",
		Value: "```Advanced Info Below```",
	}
	adv1 := &discordgo.MessageEmbedField{
		Name:   "redirectLogic",
		Value:  redirectlogic,
		Inline: true,
	}
	adv2 := &discordgo.MessageEmbedField{
		Name:   "operator",
		Value:  operator,
		Inline: true,
	}
	adv3 := &discordgo.MessageEmbedField{
		Name:   "valueToCompare",
		Value:  value,
		Inline: true,
	}
	adv4 := &discordgo.MessageEmbedField{
		Name:  "urlPart",
		Value: urlpart,
	}
	adv5 := &discordgo.MessageEmbedField{
		Name:   "validatorType",
		Value:  validatortype,
		Inline: true,
	}
	adv6 := &discordgo.MessageEmbedField{
		Name:   "isNegative",
		Value:  strconv.FormatBool(isnegative),
		Inline: true,
	}
	adv7 := &discordgo.MessageEmbedField{
		Name:   "isIgnoreCase",
		Value:  strconv.FormatBool(isignorecase),
		Inline: true,
	}
	adv8 := &discordgo.MessageEmbedField{
		Name:   "logicalOperator",
		Value:  logicaloperator,
		Inline: true,
	}
	fields = append(fields, temp1)
	fields = append(fields, temp2)
	fields = append(fields, separator)
	fields = append(fields, adv1)
	fields = append(fields, adv2)
	fields = append(fields, adv3)
	fields = append(fields, adv4)
	fields = append(fields, adv5)
	fields = append(fields, adv6)
	fields = append(fields, adv7)
	fields = append(fields, adv8)

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: domain,
			URL:  "https://" + domain + "/",
		},
		Title:       ":rotating_light: Queue-it Added To A New Product :rotating_light:",
		URL:         "https://" + domain + "/product/~/" + sku + ".html",
		Description: "A New Product has been added to Queue-it!!!!\nDescription: " + "***"+description+"***",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://images.footlocker.com/is/image/EBFL2/" + sku + "_a1",
		},
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Property of Nuggie Labs",
			IconURL: "",
		},
		Color: 14902784,
	}
	var id []*discordgo.MessageEmbed
	id = append(id, embed)
	webhook := discordgo.WebhookParams{
		Username: "Queue-it Monitor",
		Embeds:   id,
	}

	return &webhook
}

var jar, err = cookiejar.New(nil)
var client = http.Client{
	Jar: jar,
}

func main() {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	u, err := dg.User("@me")
	if err != nil {
		fmt.Println(err.Error())
	}

	BotID = u.ID

	_ = dg.Open()
	dictionary := createsplice("dictionary.txt")


	for {
		proxy, auth, password := rotateproxy(proxylist)

		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(&url.URL{
				Scheme: "http",
				User:   url.UserPassword(auth, password),
				Host:   proxy,
			}),
		}

		var current []string

		request, err := http.NewRequest("GET", "<URL HIDDEN>", nil)
		if err != nil {
			fmt.Println(err)
		}

		response, err := client.Do(request)
		if err != nil {
			fmt.Println(err)
			continue
		} else {
			defer response.Body.Close()
		}

		bodied, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println(err)
		}

		var check bod

		rest := strings.TrimPrefix(string(bodied), "window.queueit_clientside_config=")
		rest = strings.TrimSuffix(rest, ";QueueIt.Javascript.PageEventIntegration.initQueueClient(window.queueit_clientside_config);")

		err = json.Unmarshal([]byte(rest), &check)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(time.Now().String() + "\nProxy Iteration: " + strconv.Itoa(proxyiteration))

		for k := range check.Integrations {
			for trig := range check.Integrations[k].Triggers {
				for part := range check.Integrations[k].Triggers[trig].TriggerParts {
					if strings.Contains(check.Integrations[k].Triggers[trig].TriggerParts[part].Value, "html") || (len(check.Integrations[k].Triggers[trig].TriggerParts[part].Value) <= 8 && !strings.Contains(check.Integrations[k].Triggers[trig].TriggerParts[part].Value, "product")) {
						if len(check.Integrations[k].Triggers[trig].TriggerParts[part].Value) <= 1 {  		// CHECK THE ARRAY VALUESTOCOMPARE
							current, dictionary = checkarray(dg, check, k, trig, part, current, dictionary)
							continue
						}
						staging := false
						for h := range check.Integrations[k].Triggers[trig].TriggerParts {
							if strings.Contains(check.Integrations[k].Triggers[trig].TriggerParts[h].Value, "staging") {
								staging = true
							}
						}
						store := check.Integrations[k].CookieDomain
						sku := strings.TrimSuffix(check.Integrations[k].Triggers[trig].TriggerParts[part].Value, ".html")
						current = append(current, sku+":"+store) //add sku to splice
						for x := range dictionary {
							tempdic := strings.Split(dictionary[x], ":")
							if strings.Compare(tempdic[0], sku) == 0 && len(tempdic) > 1 && strings.Compare(tempdic[1], store) == 0 {
								break
							} else if x == len(dictionary)-1 {
								dictionary = append(dictionary, sku+":"+store)
								writedic(sku+":"+store, "dictionary.txt")
								fmt.Println(sku + ":" + store + " ADDED")

								var domain string
								if staging {
									domain = "staging." + check.Integrations[k].CookieDomain
								} else {
									domain = check.Integrations[k].CookieDomain
								}

								webhook := createalert(sku, domain, check.Description)
								_, _ = dg.WebhookExecute("<discord webhook>", "<discord webhook>", true, webhook)
								advwebhook := createspecialalert(sku, check.Integrations[k].Triggers[trig].TriggerParts[part].Value, domain, check.Integrations[k].RedirectLogic, check.Integrations[k].Triggers[0].LogicalOperator, check.Integrations[k].Triggers[trig].TriggerParts[part].Operator, check.Integrations[k].Triggers[trig].TriggerParts[part].UrlPart, check.Integrations[k].Triggers[trig].TriggerParts[part].ValidatorType, check.Integrations[k].Triggers[trig].TriggerParts[part].IsNegative, check.Integrations[k].Triggers[trig].TriggerParts[part].IsIgnoreCase, check.Description)
								_, _ = dg.WebhookExecute("<discord webhook>", "<discord webhook>", true, advwebhook) 
								time.Sleep(time.Second)
								//alert discord of new sku
							}
						}

					}
				}
			}
		}

		//compare dictionary to splice here
		newsplice, missing := checkdeleted(current, dictionary)
		if newsplice != nil {
			dictionary = newsplice
			//create deletionwebhook
			webhook := createdeletion(missing, check.Description)
			_, _ = dg.WebhookExecute("<discord webhook>", "<discord webhook>", true, webhook)
			//fmt.Println(dictionary)
			fmt.Println(missing + " Has Been Deleted")
		}

		time.Sleep(time.Second * 3)
	}
}
