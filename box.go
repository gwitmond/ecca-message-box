// Ecca Authentication Message Box service
//
// Lets people rent mailboxes.
// Separate user naming from mail addressing.
//
// Copyright 2013, Guido Witmond <guido@witmond.nl>
// Licensed under AGPL v3 or later. See LICENSE.

// Handle the box handling.



package main

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"crypto/rand"
	"math/big"
	"log"
	)

var DeliverDir = "/tmp/deliver"   // Place where the boxnumber symlinks to Boxdir/CN/boxnumber get stored
var BoxDir = "/tmp/box"             // Place where the boxes are stored as CN/boxnumber/
func init() {
	os.MkdirAll(DeliverDir, 0755)
	os.MkdirAll(BoxDir, 0755)
}

// if a link exists in this directory, it's a valdid box number.
// expect it to be clean numbers, free of / ..  and so on.
func boxnumberValid(mb string) bool {
	mbp := path.Join(DeliverDir, mb)
	_, err := os.Lstat(mbp)
	return  err == nil  // it exists -> true
}

// check if it is free, ignore race conditions
func boxnumberFree(mb string) bool {
	mbp := path.Join(DeliverDir, mb)
	log.Printf("Checking if %v is free\n", mbp)
	info, err := os.Lstat(mbp)
	log.Printf("lstat is %#v, err is %#v\n", info, err)
	return  info == nil  // not there: it's free
}

// getNewUniqueMB gets a new unique mailbox address
func getNewUniqueMB(cn string) string {
	mb := randBigInt().String()
	if boxnumberFree(mb) == false { 
		return getNewUniqueMB(cn) // recurse with new random number if not free
	}
	// it's free, make directory structure and symlink to it.
	boxp := path.Join(BoxDir, cn, mb)       // the directory for the messages to mb
	log.Printf("Creating dir: %v\n", boxp)
	check(os.MkdirAll(boxp, 0755))

	deliverp := path.Join(DeliverDir, mb)  // create pointer from mb to CN/mb (to ease depositing messages)
	log.Printf("Linking: %v to %v\n", boxp, deliverp)
	check(os.Symlink(boxp, deliverp))

	return mb
}

// dropoffMessage delivers a message into a messgage box.
func dropoffMessage(mb string,  messageBody io.ReadCloser) {
	// defer messageBody.Close() // caller should close
	deliverp := path.Join(DeliverDir, mb)  
	dst, err := ioutil.TempFile(deliverp, "")  
	// fname := dst.Name() // name including full path
	_, err = io.Copy(dst, messageBody)
	dst.Close() // close before checking error.
	check(err)
	// check(os.Rename(fname, fname + ".message")) //  signal it's available to avoid premature reading.
}

// deleteMessage removes the message from the storage
func deleteMessage(cn, mb, msgid string) error{
	return os.Remove(retrieveMessageFilename(cn, mb, msgid))
}

// retrieve the file containing the message. 
// As long as we prepend BoxDir and CN and forbid /../ it's safe.
// The mb and msdid are the users' responsibility.
// Notice: We assume the caller (the web interface) checks for /../
func retrieveMessageFilename(cn, mb, msgid string) string {
	return  path.Join(BoxDir, cn, mb, msgid)
}

func retrieveMessageForBox(cn, mb string) string {
	return  path.Join(BoxDir, cn, mb)
}

func retrieveMBsForCN(cn string) string {
	return  path.Join(BoxDir, cn)
}

// should go in ecca-lib
var (
        maxInt64 int64 = 0x7FFFFFFFFFFFFFFF
        maxBig64       = big.NewInt(maxInt64)
)

func randBigInt() (value *big.Int) {
        value, _ = rand.Int(rand.Reader, maxBig64)
        return
}
