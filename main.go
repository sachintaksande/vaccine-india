package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	"github.com/go-co-op/gocron"
)

func main() {

	fmt.Println("IMPORTANT:::: If you face problems with emails not being sent from the smtp configurations you have, please allow less secure apps in your account. Log in to your gmail account which you have configured and then visit https://www.google.com/settings/security/lesssecureapps and allow less secure apps.")

	var configFilePath string
	flag.StringVar(&configFilePath, "config-file", "./config.json", "path of the config.json file")

	if strings.TrimSpace(configFilePath) == "" {
		handleError(errors.New("the value of -config-file needs to be a path to a valid config.json file"))
	}
	configFilePath = strings.TrimSpace(configFilePath)
	fileAsBytes, err := os.ReadFile(configFilePath)
	handleError(err)

	config := &Configuration{}
	err = json.Unmarshal(fileAsBytes, config)
	handleError(err)

	validateNotificationConfigs(config)
	validatePollingInterval(config)
	stateDistrictMap := populateStateAndDistricts()
	validateListeners(config, stateDistrictMap)
	startListeningAndNotifying(config, stateDistrictMap)
}

func validateNotificationConfigs(config *Configuration) {
	// Only smtp support for now
	if strings.TrimSpace(config.Notificationconfigs.SMTP.Host) == "" {
		handleError(errors.New("the smtp host needs to be a valid value (notificationConfigs.smtp.host)"))
	}
	port, err := strconv.Atoi(strings.TrimSpace(config.Notificationconfigs.SMTP.Port))
	if err != nil || port <= 0 {
		handleError(errors.New("the smtp port needs to be a valid port number (notificationConfigs.smtp.port)"))
	}
	if strings.TrimSpace(config.Notificationconfigs.SMTP.Email) == "" {
		handleError(errors.New("the smtp email needs to be a valid email (notificationConfigs.smtp.email)"))
	}
	if strings.TrimSpace(config.Notificationconfigs.SMTP.Password) == "" {
		handleError(errors.New("the smtp password needs to be a valid string (notificationConfigs.smtp.password)"))
	}
}

func validatePollingInterval(config *Configuration) {
	pollingInterval, err := strconv.Atoi(strings.TrimSpace(config.Pollinginterval))
	if err != nil || pollingInterval <= 0 {
		handleError(errors.New("the polling interval needs to be a valid number (notificationConfigs.pollingInterval)"))
	}
}

func logResponse(url string, resBytes []byte) {
	// fmt.Println("============BEGIN=============")
	// fmt.Println(url)
	// fmt.Println(string(resBytes))
	// fmt.Println("============END=============")
}
func populateStateAndDistricts() map[string]map[string]int {
	statesUrl := "https://cdn-api.co-vin.in/api/v2/admin/location/states"
	res, err := makeRequest(statesUrl)
	handleError(err)
	resBytes, err := ioutil.ReadAll(res.Body)
	handleError(err)
	logResponse(statesUrl, resBytes)
	states := &CowinStates{}
	err = json.Unmarshal(resBytes, states)
	handleError(err)
	var stateMap map[string]map[string]int = make(map[string]map[string]int)
	for _, state := range states.States {
		distUrl := fmt.Sprintf("https://cdn-api.co-vin.in/api/v2/admin/location/districts/%d", state.StateID)
		res, err = makeRequest(distUrl)
		handleError(err)
		resBytes, err = ioutil.ReadAll(res.Body)
		handleError(err)
		logResponse(distUrl, resBytes)
		districts := &CowinDistricts{}
		err = json.Unmarshal(resBytes, districts)
		handleError(err)
		var districtMap map[string]int = make(map[string]int)
		for _, district := range districts.Districts {
			districtMap[strings.ToLower(district.DistrictName)] = district.DistrictID
		}
		stateMap[strings.ToLower(state.StateName)] = districtMap
	}
	return stateMap
}

func validateListeners(config *Configuration, stateDistrictMap map[string]map[string]int) {
	if len(config.Listeners) == 0 {
		handleError(errors.New("atleast one valid listner needs to be provided"))
	}
	for _, listener := range config.Listeners {
		if strings.TrimSpace(listener.Pin) != "" {
			continue
		}
		if listener.State == "" || listener.District == "" {
			handleError(errors.New("either pin or both state and district need to be valid values"))
		}
		districtMap, exists := stateDistrictMap[strings.ToLower(listener.State)]
		if !exists {
			handleError(errors.New("The value of state " + listener.State + " is invalid. Please provide a valid state name"))
		}
		_, exists = districtMap[strings.ToLower(listener.District)]
		if !exists {
			handleError(errors.New("The value of district " + listener.District +
				" is invalid. Please provide a valid district name in " + listener.State + " state"))
		}
		if len(listener.Receivers) == 0 {
			handleError(errors.New("atleast one valid receiver email needs to be provided for each listener"))
		}
		for _, email := range listener.Receivers {
			err := checkmail.ValidateFormat(email)
			handleError(err)
			err = checkmail.ValidateHost(email)
			handleError(err)
			err = checkmail.ValidateHostAndUser(config.Notificationconfigs.SMTP.Host, config.Notificationconfigs.SMTP.Email, email)
			if smtpErr, ok := err.(checkmail.SmtpError); ok && err != nil {
				handleError(errors.New("Either the receiver email in listener is invalid or the smtp configurations are incorrect. Error: " + smtpErr.Error()))
			}
		}
		if listener.Filters.MinAge != 0 && listener.Filters.MinAge != 18 && listener.Filters.MinAge != 45 {
			handleError(errors.New("invalid value in age filter. allowed values are blank, '18', '45'"))
		}
		if listener.Filters.Fees != "" && listener.Filters.Fees != "free" && listener.Filters.Fees != "paid" {
			handleError(errors.New("invalid value in fees filter. allowed values are blank, 'free', 'paid'"))
		}
		// if listener.Filters.Vaccine != "" && listener.Filters.Vaccine != "covishield" && listener.Filters.Vaccine != "covaxin" {
		// 	handleError(errors.New("invalid value in vaccine filter. allowed values are blank, 'covishield', 'covaxin'"))
		// }
		if listener.Filters.MinSlots != 0 && listener.Filters.MinSlots < 1 {
			handleError(errors.New("invalid value in minSlots filter. allowed values is greater than 0"))
		}
	}
}

func startListeningAndNotifying(config *Configuration, stateDistrictMap map[string]map[string]int) {
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.Every(getPollingInterval(config)).Seconds().Do(func() {
		for _, loopVar := range config.Listeners {
			listener := loopVar
			url := ""
			dateStr := time.Now().Local().Format("02-01-2006")
			if strings.TrimSpace(listener.Pin) != "" {
				url = fmt.Sprintf("https://cdn-api.co-vin.in/api/v2/appointment/sessions/public/calendarByPin?pincode=%s&date=%s",
					strings.TrimSpace(listener.Pin), dateStr)
			} else {
				url = fmt.Sprintf("https://cdn-api.co-vin.in/api/v2/appointment/sessions/public/calendarByDistrict?district_id=%d&date=%s",
					getDistrictCode(stateDistrictMap, listener.State, listener.District), dateStr)
			}
			res, err := makeRequest(url)
			handleError(err)
			resBytes, err := ioutil.ReadAll(res.Body)
			handleError(err)
			logResponse(url, resBytes)
			slots := &CowinSlots{}
			err = json.Unmarshal(resBytes, slots)
			handleError(err)
			var filteredSlots []Slot
			for _, center := range slots.Centers {
				for _, session := range center.Sessions {
					slotCount := 1
					if listener.Filters.MinSlots > 0 {
						slotCount = listener.Filters.MinSlots
					}
					if session.getRoundedAvailableCapacity() >= slotCount { // slots
						if listener.Filters.MinAge == 0 || session.MinAgeLimit == listener.Filters.MinAge { // age
							if listener.Filters.Fees == "" || strings.EqualFold(listener.Filters.Fees, center.FeeType) { //fees
								if listener.Filters.Vaccine == "" || strings.EqualFold(listener.Filters.Vaccine, session.Vaccine) { //vaccine
									filteredSlots = append(filteredSlots, Slot{
										Center:         fmt.Sprintf("%s,%s,%s,%s,%s,%d", center.Name, center.Address, center.BlockName, center.DistrictName, center.StateName, center.Pincode),
										AvailableSlots: fmt.Sprintf("%d", session.getRoundedAvailableCapacity()),
										Date:           session.Date,
										Vaccine:        session.Vaccine,
										Age:            fmt.Sprintf("%d", session.MinAgeLimit),
										FeeType:        center.FeeType,
									})
								}
							}
						}
					}
				}
			}
			if len(filteredSlots) == 0 {
				continue
			}
			auth := smtp.PlainAuth("", config.Notificationconfigs.SMTP.Email, config.Notificationconfigs.SMTP.Password, config.Notificationconfigs.SMTP.Host)

			t, _ := template.New("Email").Parse(EMAIL_TEMPLATE)

			var body bytes.Buffer

			mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
			body.Write([]byte(fmt.Sprintf("Subject: Covid-19 Vaccine Availability Alert \n%s\n\n", mimeHeaders)))

			t.Execute(&body, filteredSlots)
			emailBytes := body.Bytes()
			logResponse("Email Body", emailBytes)
			// Sending email.
			err = smtp.SendMail(config.Notificationconfigs.SMTP.Host+":"+config.Notificationconfigs.SMTP.Port, auth,
				config.Notificationconfigs.SMTP.Email, listener.Receivers, emailBytes)
			handleError(err)
		}
	})
	scheduler.SingletonMode()
	scheduler.StartBlocking()
}

func makeRequest(url string) (*http.Response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", url, nil)
	handleError(err)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("DNT", "1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Access-Control-Allow-Origin", "*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:88.0) Gecko/20100101 Firefox/88.0")
	return client.Do(req)
}

func getDistrictCode(stateDistrictMap map[string]map[string]int, state, district string) int {
	return stateDistrictMap[strings.ToLower(state)][strings.ToLower(district)]
}

func getPollingInterval(config *Configuration) int {
	pollingInterval, _ := strconv.Atoi(strings.TrimSpace(config.Pollinginterval))
	return pollingInterval
}

func handleError(err error) {
	if err != nil {
		panic(err.Error())
	}
}