# PNR Status Change Notifier 🚆📩  

A lightweight tool that monitors your **Indian Railways PNR status** and notifies you instantly whenever there’s a change.  
Notifications can be sent via **email**, **[ntfy.sh](https://ntfy.sh)**, or other push providers you choose.  

---

## ✨ Features  
- Track multiple PNRs at once.  
- Get notified instantly when PNR status changes (e.g., WL → CNF, berth updates, charting).  
- Supports multiple notification channels:  
  - **Email** (via GitHub Secrets).  
  - **Push notifications** using [ntfy.sh](https://ntfy.sh).  
- Runs automatically via **GitHub Actions** or locally with Go.  

---

## 🚀 Getting Started  

### 1. Fork or Clone the Repository  
```bash
git clone https://github.com/<your-username>/pnr-status-change-notifier.git
cd pnr-status-change-notifier
``` 

### 2. Configure GitHub Secrets

Go to your repo → Settings → Secrets and variables → Actions → New repository secret.

## Add these secrets (comma-separated lists, aligned by position):
-	PNR_LIST → List of PNR numbers
-	Example: 1234567890,2345678901
-	EMAIL_LIST → List of emails (each mapped to corresponding PNR)
-	Example: user1@example.com,user2@example.com
-	QUEUE_LIST → List of ntfy.sh queue names (each mapped to corresponding PNR)
-	Example: pnr1234,pnr5678

Each entry in the lists must match in order.

	
### 3.	Run the program:

```bash
go run main.go
``` 
