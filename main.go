package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Response struct {
	Data struct {
		PnrResponse struct {
			Pnr             string `json:"pnr"`
			TrainNo         string `json:"trainNo"`
			TrainName       string `json:"trainName"`
			Doj             string `json:"doj"`
			From            string `json:"from"`
			To              string `json:"to"`
			Class           string `json:"class"`
			ChartPrepared   bool   `json:"chartPrepared"`
			CacheTime       string `json:"cacheTime"`
			PassengerStatus []struct {
				Number        int    `json:"number"`
				BookingStatus string `json:"bookingStatus"`
				CurrentStatus string `json:"currentStatus"`
				Coach         string `json:"coach"`
				Berth         int    `json:"berth"`
			} `json:"passengerStatus"`
		} `json:"pnrResponse"`
	} `json:"data"`
}

func getPNRResponse(pnr string) Response {
	url := fmt.Sprintf("https://cttrainsapi.confirmtkt.com/api/v2/ctpro/mweb/%s?querysource=ct-web&locale=en&getHighChanceText=true&livePnr=false", pnr)

	payload := strings.NewReader(`{"proPlanName":"CP7","emailId":"","tempToken":""}`)

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		panic(err)
	}

	// Add required headers
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8,hi;q=0.7")
	req.Header.Add("ApiKey", "ct-web!2$")
	req.Header.Add("CT-Token", "")
	req.Header.Add("CT-Userkey", "")
	req.Header.Add("ClientId", "ct-web")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Origin", "https://www.confirmtkt.com")
	req.Header.Add("Referer", "https://www.confirmtkt.com/")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		panic(err)
	}
	return result
}

func getPassengersCurrentStatus(response Response) string {
	currentStatus := ""
	for _, p := range response.Data.PnrResponse.PassengerStatus {
		currentStatus += fmt.Sprintf("#%d->Current: %s Berth: %d Coach: %s",
			p.Number, p.CurrentStatus, p.Berth, p.Coach)
	}
	return currentStatus
}

func readLastPNRData(pnr string) (string, error) {
	if len(pnr) < 4 {
		return "", fmt.Errorf("invalid PNR: must have at least 4 digits")
	}

	filename := fmt.Sprintf("%s.txt", pnr[len(pnr)-4:])

	filepath := filepath.Join(".", filename)

	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		f, err := os.Create(filepath)
		if err != nil {
			return "", fmt.Errorf("failed to create file: %v", err)
		}
		f.Close()
		return "", nil // return empty since new file
	}

	// Read existing file content
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	return string(data), nil
}

func writePNRData(pnr string, data string) error {
	if len(pnr) < 4 {
		return fmt.Errorf("invalid PNR: must have at least 4 digits")
	}

	filename := fmt.Sprintf("%s.txt", pnr[len(pnr)-4:])
	filepath := filepath.Join(".", filename)

	// Write data (overwrite mode)
	err := ioutil.WriteFile(filepath, []byte(data), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}
	return nil
}

func sendEmailAndPushNotification(uri, message, email string) {
	req, _ := http.NewRequest("POST", uri,
		strings.NewReader(message))
	req.Header.Set("Email", email)
	req.Header.Set("Priority", "high")
	do, err := http.DefaultClient.Do(req)
	defer do.Body.Close()
	if err != nil || do.StatusCode != 200 {
		fmt.Println("Failed to send email notification!")
	}
	fmt.Println("Successfully sent email notification!")
}

func sendPushNotification(uri, message string) {
	post, err := http.Post(uri, "text/plain",
		strings.NewReader(message))
	defer post.Body.Close()
	if err != nil || post.StatusCode != 200 {
		fmt.Println("Failed to send push notification!")
	}
	fmt.Println("Successfully sent push notification!")
}

func main() {
	pnrs := os.Getenv("PNR_LIST")
	pnrList := strings.Split(pnrs, ",")

	ReciepientEmails := os.Getenv("RECIPIENT_EMAILS")
	emailList := strings.Split(ReciepientEmails, ",")

	notifQueues := os.Getenv("NOTIF_QUEUES")
	notifQueueList := strings.Split(notifQueues, ",")

	notifServiceProvider := os.Getenv("NOTIF_SERVICE_PROVIDER")

	for idx, pnr := range pnrList {
		resp := getPNRResponse(pnr)

		currentStatus := getPassengersCurrentStatus(resp)
		lastStatus, _ := readLastPNRData(pnr)

		if !strings.Contains(lastStatus, currentStatus) {
			fmt.Println("Status Changed: sending notification!")

			uri := fmt.Sprintf("%s%s", notifServiceProvider, notifQueueList[idx])
			message := fmt.Sprintf("TrainName: %s Status: %s", resp.Data.PnrResponse.TrainName, currentStatus)

			sendPushNotification(uri, message)
			sendEmailAndPushNotification(uri, message, emailList[idx])
		}


		// Load IST timezone
		loc, err := time.LoadLocation("Asia/Kolkata")
		if err != nil {
			fmt.Printf("Error loading IST timezone: %v\n", err)
			return
		}
	
		// Get current IST time in human-readable format
		currentTime := time.Now().In(loc).Format("2006-01-02 15:04:05")
		// save current status
		fileData := fmt.Sprintf("%s_CacheTime: %s_CheckedAt: %s", currentStatus, resp.Data.PnrResponse.CacheTime, currentTime)
		if err := writePNRData(pnr, fileData); err != nil {
			fmt.Println(err)
		}
		fmt.Println("Processed PNR1")
	}
}
