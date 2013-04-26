// Ecca Authentication Message Box service
//
// Lets people rent mailboxes.
// Separate user naming from mail addressing.
//
// Copyright 2013, Guido Witmond <guido@witmond.nl>
// Licensed under AGPL v3 or later. See LICENSE.

package main

import (
        "log"
        "net/http"
        "crypto/tls"
        "html/template"
	"regexp"

	"github.com/gwitmond/ecca-lib"
)

var siteURL = "https://messagebox.wtmnd.nl:1443"

var ecc = ecca.Ecca{
	RegisterURL:  "https://register-messagebox.wtmnd.nl:1444/register-pubkey",
	RegisterTemplate: template.Must(template.ParseFiles("needToRegister.template", "menu.template", "tracking.template")),
}

var homePageT = template.Must(template.ParseFiles("homepage.template", "menu.template", "tracking.template")) 
var accountPageT = template.Must(template.ParseFiles("accountPage.template", "menu.template", "tracking.template")) 
var accountCreateT = template.Must(template.ParseFiles("accountCreate.template", "menu.template", "tracking.template")) 

var mailboxPageT = template.Must(template.ParseFiles("mailboxPage.template", "menu.template", "tracking.template")) 
var mailboxCreateT = template.Must(template.ParseFiles("mailboxCreate.template", "menu.template", "tracking.template")) 

var messagesPageT = template.Must(template.ParseFiles("messagesPage.template", "menu.template", "tracking.template")) 
var messageCreateT =  template.Must(template.ParseFiles("messageCreate.template", "menu.template", "tracking.template")) 
var messageDeliverT =  template.Must(template.ParseFiles("messageDeliver.template", "menu.template", "tracking.template")) 

var messageRetrieveT = template.Must(template.ParseFiles("messageRetrieve.template", "menu.template", "tracking.template")) 

func main() {
        http.HandleFunc("/", homePage)

        http.HandleFunc("/account", accountPage)
	http.HandleFunc("/create-account", accountCreate)

	http.HandleFunc("/mailbox", mailboxPage)
	http.HandleFunc("/create-mailbox", mailboxCreate)
	//http.HandleFunc("/destroy-mailbox", mailboxDestroy)

	http.HandleFunc("/messages", messagesPage)

	http.HandleFunc("/deliver/", messageDeliver)
	http.HandleFunc("/retrieve/", messageRetrieve)
	
        http.Handle("/static/", http.FileServer(http.Dir(".")))

        // This CA-pool specifies which client certificates can log in to our site.
        pool := ecca.ReadCert("messageboxFPCA.cert.pem")
        
        log.Printf("Starting. Please visit %v", siteURL)
	server6 := &http.Server{Addr: "messagebox.wtmnd.nl:1443",
                TLSConfig: &tls.Config{
                        ClientCAs: pool,
                        ClientAuth: tls.VerifyClientCertIfGiven},
        }
        check(server6.ListenAndServeTLS("messagebox.wtmnd.nl.cert.pem", "messagebox.wtmnd.nl.key.pem"))
}

func homePage(w http.ResponseWriter, req *http.Request) {
	// Test for / explicit or we will run for every request that is not handled by other handlers.
        if req.URL.Path == "/" {
                err := homePageT.Execute(w, nil) 
                check(err)
                return
        }
        http.NotFound(w, req) // 404 - not found
}

// Show current accounts or an offer to create one.
func accountPage(w http.ResponseWriter, req *http.Request) {
	cn := "Stranger"
	if len(req.TLS.PeerCertificates) == 1 {
		cn = req.TLS.PeerCertificates[0].Subject.CommonName
        }
	check(accountPageT.Execute(w, map[string]interface{}{
                        "CN": cn,
        }))
	return
}

func accountCreate(w http.ResponseWriter, req *http.Request) {
	// Require to be logged in.
	if len(req.TLS.PeerCertificates) == 0 {
		ecc.SendToLogin(w)
                return
        }
	switch req.Method {
        case "GET": 
		cn := req.TLS.PeerCertificates[0].Subject.CommonName
		err := accountCreateT.Execute(w, map[string]interface{}{
                        "CN": cn,
                }) 
                check(err)
		return  
		
        default: 
		w.Write([]byte("Unexpected method"))
        }
        return
}

// Show current accounts or an offer to create one.
func mailboxPage(w http.ResponseWriter, req *http.Request) {
	check(mailboxPageT.Execute(w, nil))
}

// mailboxCreate creates a new mailbox and registers it with the current logged in certificate user.
func mailboxCreate(w http.ResponseWriter, req *http.Request) {
	// Require to be logged in.
	if len(req.TLS.PeerCertificates) == 0 {
		ecc.SendToLogin(w)
                return
        }

	// Todo: validate policy wrt # of mailboxes, total size...

	// create new mailbox
	cn := req.TLS.PeerCertificates[0].Subject.CommonName
	mb := getNewUniqueMB(cn)

	// show it
	check(mailboxCreateT.Execute(w, map[string]interface{}{
                "CN": cn,
		"mb": mb,
		"site": siteURL,
        }))
	return
}

// Show message page 
func messagesPage(w http.ResponseWriter, req *http.Request) {
	check(messagesPageT.Execute(w, nil))
}

var mbRE  = regexp.MustCompile(`^/deliver/([\d]+)$`)

func messageDeliver(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
        case "GET": 
		err := messageCreateT.Execute(w, nil)   // just show a form to post
                check(err)
                return

	case "POST":
		// get and validate message box number
		path := req.URL.Path
		mb := getFirst(mbRE.FindStringSubmatch(path))
		if mb == "" {
			http.Error(w, "There is no boxnumber to deposit into.", 400)
			return
		}

		// Check if the box exists. Return here before reading potential large message body...
		if boxnumberValid(mb) == false {
			http.NotFound(w, req)
			return
		}

		// deposit message
		defer req.Body.Close()
		dropoffMessage(mb, req.Body)

		// report success
		check(messageDeliverT.Execute(w, nil))
	}
}

// URL is /retrieve/<boxnumber>
// or        /retrieve/<boxnumber>/<messageid>
var mbretRE  = regexp.MustCompile(`^/retrieve/([\d]+)?/?([\d]+)?$`)

func messageRetrieve(w http.ResponseWriter, req *http.Request) {
	// Require to be logged in.
	if len(req.TLS.PeerCertificates) == 0 {
		ecc.SendToLogin(w)
                return
        }

	path := req.URL.Path
	match := mbretRE.FindStringSubmatch(path)
	if match == nil { 
		http.Error(w, "Could not parse boxnumber and message-id.", 400)
		return
	}
	cn := req.TLS.PeerCertificates[0].Subject.CommonName

	switch req.Method {
        case "GET": 
		switch {
		case match[1] == "":
			// Show all messageboxes for CN
			http.ServeFile(w, req,retrieveMBsForCN(cn))
		case match[2] == "":
			// Show messages for CN/boxnumber
			http.ServeFile(w, req, retrieveMessageForBox(cn, match[1]))
		default: 
			// Show message boxnumber/message-id
			http.ServeFile(w, req, retrieveMessageFilename(cn, match[1], match[2]))
		}

	case "DELETE":
		if match[1] == "" || match[2] == "" {
			http.Error(w, "Need a full path to a message to delete.", 400)
			return
		}
		check(deleteMessage(cn, match[1], match[2]))
		w.Write([]byte("Deleted\n"))
		return

	default: 
		http.Error(w, "Unexpected method", 400)
	}
	
}

func check(err error) {
        if err != nil {
                panic(err) // TODO: change panic to 500-error.
        }
}


// Return the first (not zeroth) string in the array, if not nil
func getFirst(s []string) string {
        if s != nil {
		return s[1]
        }
        return ""       
}

