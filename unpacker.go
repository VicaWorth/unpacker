/*
/TODO
►	1. Add the purpose.
	2. Test the usage to make sure it works.
►	3. Note the Input Flags
►	4. Note the Output Codes
	5. Remove notes and clean up comments
►	6. Just use t and not Now
►	7. Do not write over files in trash or imported. Add a random number to their file name
►	8. Is it possible for a file to be something else beside IsDirectory or IsRegular? Turn into If Else
►	9. One code per error
►	10 Can this be turned into a function or one line:
			New := fmt.Sprintf("Error with command line argument 2 bank folder: %v", err)
			Newerr := errors.New(New)
			ErrStop(2, Newerr, t)
►	11. change file name from 250.CloudCoins.word.16666533.stack to 250.CloudCoins.1.16666533.e385720bf37b48dab4d5e8aff5ca0d74.stack (1 is the network number and the GUID is the receipt ID)
►	12. Log file must be in the Logs\Unpacker folder
►	13. bank server in the receipt file will change from this:  "bank_server": "ccc.cloudcoin.global", to this: "bank_server": "localhost",
►	14. Change the name of the receipt file to 	84a5973506374f759d5bb79c2c5a14d1.txt where the first part is the receipt number.
►	15. Change the statuses in the receipt to:

			"status": "suspect",
			"pown": "uuuuuuuuuuuuuuuuuuuuuuuuu",
			"note": "Moved to the Suspect Folder"
		},


			"status": "duplicate",
			"pown": "uuuuuuuuuuuuuuuuuuuuuuuuu",
			"note": "Already in the Vault Folder"
		},
►	16. Success code is 100
	17: Receipt change: "total_suspect": 0, to "total_suspect": 340 so that the number of coins are count. Also include a count of previously imported. and a total_unchecked that can be set to zero.
			"total_unchecked": 0,
			"prev_imported": 1,
			"total_suspect": 0,
	18: Rename foldernames to better fit the program.
	19: Update all online documentation

unpacker.go
Version: 2019-8-08
Author(s): Victoria Worthington
Purpose:  To take stack file(s) and split them up into their individual coins, then check in bank, fracked, gallery, lost, & vault.
Reference: https://cloudcoinconsortium.org/software.html
Usage: This program is intended to be used for taking stack(s) of CloudCoins and seperating them, placing them into suspect or trash.
unpacker.exe -tag=word -rootpath=C:\Users\victo\Documents\CloudCoin\Accounts\MyAccount\ -toFolder=\suspect\ -fromFolder=\import\

Input Flags:
@rootpath: The path to the folder that contains the other folders such as Bank, Fracked, Log etc.
@toFolder: The path to the folder that coins will be sent, usually Suspect.
@fromFolder: the path to the folder that the coins will come from, usually Import
@tag: The tag that will be added to the coins.

Output Codes:
1: Could not read from the fromFolder.
2: No file was in the fromFolder, it was empty
3: There was an error getting file information.
4: It is not an mode.IsRegular type
5: Could not switch on the fi.Mode
6: You are missing the rootpath Flag. It should look like: (Should place drive location before aswell.) \CloudCoin\Accounts\NameOfAccount\
7: You are missing the fromFolder Flag. It should look like literal: \import\
8: You are missing the tag Flag. It should look like: string
9: You are missing the toFolder Flag. It should look like literal: \suspect\
10: Path to bank does not exist
11: file type not supporte by unpacker
12: Unable to move file, usually because file name was not added at the end.
13: Error with os.Open
14: Missing serial number in cloudcoin
15: Missing authentication number in cloudcoin
16: Missing serial number in cloudcoin, singular
17: Missing authentication number in cloudcoin, singular
18: Could not create cloudcoin in toFolder
19: No cloudcoins to make a receipt of
20: error was made on receipt checker, index and cccount could not work together
21: Could not create receipt
22: Error opening logs
23: Error writing in logs
24: Error writing in logs
25: Error writing in logs
26: The logging process failed.
27: failed to read the folder, gallery, lost, bank, etc.
28:
100: 100% good!
Copywrite: RAIDA Tech, All Rights Reserved
*/

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

//G L O B A L  S T R U C T U R E S

//Stack struct which contains an array of CloudCoins
type Stack struct {
	CloudCoin []CloudCoin `json:"cloudcoin"`
}

//CloudCoin User struct which contains a namea type and a list of social links
type CloudCoin struct {
	NN  string   `json:"nn"`
	SN  string   `json:"sn"`
	ANs []string `json:"an"`
}

//ReceiptInfo struct that collects information for receipt
type ReceiptInfo struct {
	Status         []string
	SNs            []string
	Denom          []string
	Note           []string
	CloudCoinCount []int
}

// G L O B A L    V A R S
var loc = time.FixedZone("UTC", 0)
var t = time.Now()
var date = fmt.Sprintf(t.Format("2006-01-02"))
var batchNumber = GenGUI()

var receiptinfo ReceiptInfo

var fileName string
var rootpath string
var fromFolder string
var tag string
var toFolder string

var status []string

//var recieverSN string

// var intOnes int
// var intFives int
// var intTwentyfives int
// var intHundreds int
// var intTwohundredfifties int

func main() {

	// All servants will need a function to load their flags into the global variables
	LoadFlags()

	// All servants will need to validate their inputs
	ValidateInput()

	//var stack Stack
	var suspectFolder []string
	var filesToUnpack []string
	var stack Stack

	bank := `\bank\`
	fracked := `\fracked\`
	gallery := `\gallery\`
	lost := `\lost\`
	vault := `\vault\`

	fromFolder = rootpath + fromFolder
	toFolder = rootpath + toFolder

	//filesToUnpack takes files name and puts it into an array
	importFolder, err := ioutil.ReadDir(fromFolder)
	if err != nil {
		ErrStop(1, err, t)
	}

	if len(importFolder) == 0 {
		ErrStop(2, errors.New("No file in import folder"), t)
	} else {
		for _, file := range importFolder {
			fi := fromFolder + file.Name()
			filesToUnpack = append(filesToUnpack, fi)
		} //end for
	} //end of if

	for i := range filesToUnpack {

		//gets info on file
		fi, err := os.Stat(filesToUnpack[i])

		//checks for error
		if err != nil {
			ErrStop(3, err, t)
		} //emd of if

		//switches on if the file is a regular or not
		switch mode := fi.Mode(); {
		case mode.IsRegular():
			fileExt := FileTypeChecker(filesToUnpack[i])
			if fileExt == ".stack" {
				var fileToUnpack []string
				fileToUnpack = append(fileToUnpack, filesToUnpack[i])

				//Counts and adds how many files there are to the cloudcoin count (appending them)
				receiptinfo.CloudCoinCount = append(receiptinfo.CloudCoinCount, 0)
				stack, suspectFolder = StackUnpacker(fileToUnpack, fileExt, i)
			} else {
				//Is not a regular filetype
				var err error
				ErrStop(4, err, t)
			} //end of else

			//takes bank location and the suspectFolder's file names
			DuplicateChecker(rootpath+bank, suspectFolder, "Bank Folder", toFolder)
			DuplicateChecker(rootpath+fracked, suspectFolder, "Fracked Folder", toFolder)
			DuplicateChecker(rootpath+gallery, suspectFolder, "Gallery Folder", toFolder)
			DuplicateChecker(rootpath+lost, suspectFolder, "Lost Folder", toFolder)
			DuplicateChecker(rootpath+vault, suspectFolder, "Vault Folder", toFolder)

			//num := TrashedNumGenerator()
			//MoveFile(filesToUnpack[i], rootpath+`\imported\`+num+"."+fi.Name())

			// fmt.Println("\n i: ", i)
			// fmt.Println("CC count: ", receiptinfo.CloudCoinCount)
			// fmt.Println("Denom: ", receiptinfo.Denom)
			// fmt.Println("Note: ", receiptinfo.Note)
			// fmt.Println("SNs: ", receiptinfo.SNs)
			// fmt.Println("Status: ", receiptinfo.Status)
			// fmt.Println("\n")

			fileName := fi.Name()
			LogCoinMove("No Error", fileName, 3)
		default:
			ErrStop(5, err, t)
		} //emd of siwtch

	} //end of loop
	//writes receipt
	ReceiptWriter(stack)

	//ends program
	ErrStop(100, errors.New("100 no errors"), t)
} //end of main

// A S P E C T   F U N C T I O N S

//LoadFlags makes sure that all the flags are there and warns if they are missing.
//used name return values when adding additional flags
func LoadFlags() {

	errorMessage := "There was an error with your flags: "

	flag.StringVar(&rootpath, "rootpath", "", "Path the the users wallet folder within the account folder.")
	flag.StringVar(&toFolder, "toFolder", "", "The path to the suspect folder or other")
	flag.StringVar(&fromFolder, "fromFolder", "", "The path to the import folder or other")
	flag.StringVar(&tag, "tag", "", "The memo of the coins being sent")
	flag.Parse()

	if rootpath == "" {
		ErrStop(6, errors.New(errorMessage+`You are missing the rootpath Flag. It should look like: (Should place drive location before aswell.) \CloudCoin\Accounts\NameOfAccount\`), t)
	}

	if fromFolder == "" {
		ErrStop(7, errors.New(errorMessage+`You are missing the fromFolder Flag. It should look like literal: \import\`), t)
	}

	if tag == "" {
		ErrStop(8, errors.New(errorMessage+`You are missing the tag Flag. It should look like: string`), t)
	}

	if toFolder == "" {
		ErrStop(9, errors.New(errorMessage+`You are missing the toFolder Flag. It should look like literal: \suspect\`), t)
	}

} //end load flags

//ValidateInput looks flags to see if they meet the specifications
func ValidateInput() {
	/* Validate inputs */

	//Validate bank path
	if _, err := os.Stat(rootpath + `\bank\`); os.IsNotExist(err) {
		fmt.Print(rootpath + `\bank\`)
		// path/to/bank does not exist
		ErrStop(10, errors.New("error with command line argument 2 bank folder"), t)
	} //end if

} //end Validate Inputer

// H E L P E R  F U N C T I O N S

//FileTypeChecker checks if file is a png, jpg, gif, stack. or other types
func FileTypeChecker(filesToUnpack string) string {
	fileType := path.Ext(filesToUnpack)
	var fileExt string
	//Switches on type (will add modules for each later)
	switch fileType {
	//case ".png":
	//	fileExt = ".png"
	//case ".jpg":
	//	fileExt = ".jpg"
	//case ".gif":
	//	fileExt = ".gif"
	case ".stack":
		fileExt = ".stack"
	case ".chest":
		fileExt = ".stack"
	default:
		ErrStop(11, errors.New("file type not supporte by unpacker"), t)
		fileExt = "Not Recongnized"
	} //end of switch
	return fileExt
} //end of FileTypeChecker

//MoveFile to a different location (used for sent coins)
func MoveFile(sourcePath, destPath string) {
	err := os.Rename(sourcePath, destPath)
	if err != nil {
		ErrStop(12, err, t)
	} //end of if

} //end of movefile func

//StackUnpacker takes a stack and breaks it up
func StackUnpacker(filesToUnpack []string, fileType string, index int) (Stack, []string) {
	//loops through files to Unpack
	var nn []string
	var sn []string
	var an [][]string
	var stack Stack

	var suspectPath []string

	for j, fileName := range filesToUnpack {
		jsonFile, err := os.Open(fileName)

		// if we os.Open returns an error then handle it
		if err != nil {
			ErrStop(13, err, t)
		}
		// read our opened json as a byte array.
		byteValue, _ := ioutil.ReadAll(jsonFile)

		// we unmarshal our byteArray which contains our
		// jsonFile's content into 'users' which we defined above
		json.Unmarshal(byteValue, &stack)

		checker := string(byteValue)
		coinCount := strings.Count(checker, "pown")

		//build the arrays for the Send Request
		if strings.Count(checker, "pown") >= 2 { //checks if there is more than one coin in the file
			for f := 0; f < coinCount; f++ {
				nn = append(nn, stack.CloudCoin[0].NN)
				sn = append(sn, stack.CloudCoin[0].SN)

				for i := 0; i < 25; i++ {
					if i == 0 {
						an = append(an, []string{string(stack.CloudCoin[0].ANs[i])})
					} else {
						//an = append(an, []string{""})
						an[j] = append(an[j], stack.CloudCoin[0].ANs[i])
					} //end of else
				} //end for every an

				//Checks if there is an SN in the CC
				if stack.CloudCoin[f].SN == "" {
					ErrStop(14, fmt.Errorf("missing sn(serial number) in cloudcoin %s- please remove the file or fix the issue", stack.CloudCoin[f].SN), t)
				} //end of if statement

				for file := range stack.CloudCoin[f].ANs {
					for i := 0; i > 24; i++ {
						if stack.CloudCoin[f].ANs[i] == "" {
							ErrStop(15, fmt.Errorf("Missing AN(authentication number) in cloudcoin %d- an index number %s- please remove the file or fix the issue", file, stack.CloudCoin[f].ANs[i]), t)
						} //end of if statement
					} //end of internal for loop
				} //end of external loop

				//file gets what file just got sent to suspect
				file := FilesWriter(stack, f, toFolder, fileName, fileType)
				suspectPath = append(suspectPath, file)
				LogCoinMove("No Error", stack.CloudCoin[f].SN, 1)
				jsonFile.Close()

				//adds how many cloudcoins to the cloudcoin receipt counter
				receiptinfo.CloudCoinCount[index] = receiptinfo.CloudCoinCount[index] + 1
			} //end of for loop for multiple coins in a file
		} else {
			nn = append(nn, stack.CloudCoin[0].NN)
			sn = append(sn, stack.CloudCoin[0].SN)

			//loops and adds each an
			for i := 0; i < 25; i++ {
				if i == 0 {
					an = append(an, []string{string(stack.CloudCoin[0].ANs[i])})
				} else {
					an[j] = append(an[j], stack.CloudCoin[0].ANs[i])
				} //end of else
			} //end for every an

			//Checks if there is an SN in the CC
			if stack.CloudCoin[0].SN == "" {
				ErrStop(16, fmt.Errorf("Missing SN(serial number) in cloudcoin %s- please remove the file or fix the issue", fileName), t)
			} //end of if

			//Checks if there was a missing AN in the CC
			for file := range stack.CloudCoin[0].ANs {
				for i := 0; i > 24; i++ {
					if stack.CloudCoin[0].ANs[i] == "" {
						ErrStop(17, fmt.Errorf("Missing AN(authentication number) in cloudcoin %d- an index number %s- please remove the file or fix the issue", file, stack.CloudCoin[0].ANs[i]), t)
					} //end of if statement
				} //end of internal for loop
			} //end of external loop

			//file gets what file just got sent to suspect
			file := FilesWriter(stack, 0, toFolder, fileName, fileType)
			suspectPath = append(suspectPath, file)

			LogCoinMove("No Error", stack.CloudCoin[0].SN, 1)

			//adds how many cloudcoins to the cloudcoin receipt counter
			receiptinfo.CloudCoinCount[index] = receiptinfo.CloudCoinCount[index] + 1
			jsonFile.Close()

		} // end for each file to send
	} //end of else
	return stack, suspectPath
} //end of func

//FilesWriter creates the file and puts the CloudCoin into it.
func FilesWriter(stack Stack, f int, suspectFolder, fileName, fileType string) string {
	//var stack Stack
	Part1 := fmt.Sprintf(
		`{
	"cloudcoin":
	[
		{
		"nn":"%s",
		"sn":"%s",`, stack.CloudCoin[f].NN, stack.CloudCoin[f].SN) //end of part1

	Part2 := fmt.Sprintf(
		`
		"an": ["%s", "%s", "%s", "%s", "%s",
		"%s", "%s", "%s", "%s", "%s",
		"%s", "%s", "%s", "%s", "%s",
		"%s", "%s", "%s", "%s", "%s",
		"%s", "%s", "%s", "%s", "%s"]`, stack.CloudCoin[f].ANs[0], stack.CloudCoin[f].ANs[1], stack.CloudCoin[f].ANs[2], stack.CloudCoin[f].ANs[3], stack.CloudCoin[f].ANs[4],
		stack.CloudCoin[f].ANs[5], stack.CloudCoin[f].ANs[6], stack.CloudCoin[f].ANs[7], stack.CloudCoin[f].ANs[8], stack.CloudCoin[f].ANs[9],
		stack.CloudCoin[f].ANs[10], stack.CloudCoin[f].ANs[11], stack.CloudCoin[f].ANs[12], stack.CloudCoin[f].ANs[13], stack.CloudCoin[f].ANs[14],
		stack.CloudCoin[f].ANs[15], stack.CloudCoin[f].ANs[16], stack.CloudCoin[f].ANs[17], stack.CloudCoin[f].ANs[18], stack.CloudCoin[f].ANs[19],
		stack.CloudCoin[f].ANs[20], stack.CloudCoin[f].ANs[21], stack.CloudCoin[f].ANs[22], stack.CloudCoin[f].ANs[23], stack.CloudCoin[f].ANs[24]) //end of part2

	Part3 := fmt.Sprintf(
		`
		"ed": "",
		"pown":"uuuuuuuuuuuuuuuuuuuuuuuuu",
		"aoid": []
		}
	]
}
	`) //end of part3

	//Sets up information for message
	sn, _ := strconv.Atoi(stack.CloudCoin[f].SN)
	ammountOfCoins := strconv.Itoa(Denomination(sn))
	suspectFolder = rootpath + `\suspect\`
	fileName = ammountOfCoins + ".CloudCoins." + stack.CloudCoin[f].NN + "." + batchNumber + "." + stack.CloudCoin[f].SN + fileType

	//appends information to receiptinfo, will later inspect and see if fileLocation changes
	receiptinfo.SNs = append(receiptinfo.SNs, stack.CloudCoin[f].SN)
	receiptinfo.Status = append(receiptinfo.Status, "suspect")
	receiptinfo.Denom = append(receiptinfo.Denom, ammountOfCoins)
	receiptinfo.Note = append(receiptinfo.Note, "Moved to the Suspect Folder")

	//concatinates and turns vars into bytes
	CloudCoin1 := []byte(Part1 + Part2 + Part3)

	var _, err = os.Stat(suspectFolder)

	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(suspectFolder)
		if err != nil {
			ErrStop(18, err, t)
		}
		defer file.Close()
	} //end of if

	suspectFolder = suspectFolder + fileName
	ioutil.WriteFile(suspectFolder, CloudCoin1, 0644)

	fileName = fileName + ""

	return fileName
} //end of function

//LogCoinMove logs whenever a coin moves into the unpacker.txt
func LogCoinMove(errLog, coinSN string, logType int) {

	newline := fmt.Sprintf("\n")
	//will make a receipt and an unpacker log
	logsFolder := rootpath + `\logs\` + `unpacker\` + date + `.unpacker` + ".txt"
	//checks if logs folder exists or not, if not it will make it
	if _, err := os.Stat(logsFolder); os.IsNotExist(err) {
		os.Create(logsFolder)
	} //end of thing

	//gets current time and rounds it, then makes message

	message := t.String() //+ loc.String() + ",[UNPACKER],"

	f, err := os.OpenFile(logsFolder,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		ErrStop(22, err, t)
	}
	//switches based on what happened to the coin.
	switch logType {
	//File moved to suspect
	case 1:
		message = message + `[UNPACKED], [` + coinSN + "], [" + errLog + "]" + newline
		defer f.Close()
		if _, err := f.WriteString(message); err != nil {
			ErrStop(23, err, t)
		}
	//Duplicate found
	case 2:
		message = message + `[DUPLICATED], [` + coinSN + "], [" + errLog + "]" + newline
		defer f.Close()
		if _, err := f.WriteString(message); err != nil {
			ErrStop(24, err, t)
		} //end of error check
		//Incorrect Number
	case 3:
		message = message + `[FOLDERS], [` + errLog + `]` + newline
		defer f.Close()
		if _, err := f.WriteString(message); err != nil {
			ErrStop(25, err, t)
		} //end of error check
	default:
		//something went wrong in the actual logging process
		ErrStop(26, fmt.Errorf("the logging process failed"), t)
	} //end of switch

} //end of func LogCoinMove

//DuplicateChecker looks through files and sees if there are any duplicate coins.
func DuplicateChecker(folderPath string, CCs []string, folderName, suspectPath string) {
	//the suspect folder's SNs
	var suspectSNs []string

	//Has the full name of each file in the folder
	var fileNames []string

	//bSNs is each file's sn number within folders other than suspect.
	var bSNs []string

	//Trash path leads to the trashfolder
	trashPath := rootpath + `\trash\`

	//folderContent is a file os type, is used to get information from any folder besides suspect.
	folderContent, err := ioutil.ReadDir(folderPath)
	if err != nil {
		ErrStop(27, err, t)
	}
	if len(folderContent) == 0 {
	} else {
		//ranges through that folder and starts appending file names, as well as getting SN numbers.
		for i, file := range folderContent {
			fileNames = append(fileNames, file.Name())
			//sn represents the content between the dots in a name, currently set to find contents after 4 dots.
			sn := strings.Split(fileNames[i], ".")
			bSNs = append(bSNs, sn[4])
		} //end for

		for i := range CCs {
			//sn represents the content between the dots in a name, currently set to find contents after 4 dots.
			sn := strings.Split(CCs[i], ".")
			suspectSNs = append(suspectSNs, sn[4])
		} //end of for loop

		//loops through the bank's sns, then looping through the those CCs. Checks if the name contains the SN.
		for _, bSN := range bSNs {
			for i, suspectSN := range suspectSNs {
				if strings.Contains(bSN, suspectSN) {
					errLog := fmt.Sprintf("Duplicate file found in %s, moving file to trash...", folderName)

					//deleting file
					num := TrashedNumGenerator()
					suspectCoinPath := suspectPath + CCs[i]
					trashCoinPath := trashPath + num + "." + CCs[i]

					//checks if the file exists
					if _, err := os.Stat(suspectCoinPath); os.IsNotExist(err) {
						// path/to/whatever does not exist
						if err != nil {
							err := fmt.Sprintf("path to %s does not exist, duplicate checker, possibly already processed", suspectCoinPath)
							LogCoinMove(err, suspectSN, 1)
						} //end of outer else
					} else {
						MoveFile(suspectCoinPath, trashCoinPath)
						//logging the move
						LogCoinMove(errLog, suspectSN, 2)

						//collects information for receipt. Checks if suspectSN is equal to any existing SNs, if so it will replace the filelocation to its new place.
						for j := range receiptinfo.CloudCoinCount {
							if suspectSN == receiptinfo.SNs[j] {
								receiptinfo.Status[i] = "duplicate"
								receiptinfo.Note[i] = "Already in " + folderName
							} //end of if

						} //end of loop

					} // end of else

				} //end of if statement

			} //end of suspectSns loop

		} //end of bsns loop

	} //end of else statement
} //end of func

//ReceiptWriter creates receipts.
func ReceiptWriter(stack Stack) {
	//generates starting string
	var receipt string

	//len of receiptinfo.CloudCoinCount is the quantity of files

	//checks if there are cloudcoins, otherwise it will not make a receipt
	fmt.Println(receiptinfo.CloudCoinCount)
	if len(receiptinfo.CloudCoinCount) == 0 {
		ErrStop(19, fmt.Errorf("there were no cloudcoins to make a receipt of"), t)
	} else {
		//start of receipt
		receipt1 := fmt.Sprintf(
			`{
		"receipt_id": "%s",
		"time": "%s",
		"timezone": "%v",
		"bank_server": "localhost",
		"total_authentic": %d,
		"total_fracked": %d,
		"total_counterfeit": %d,
		"total_lost": %d,
		"total_suspect": %d,
		"receipt_detail": [{`, "Word", t, loc, 0, 0, 0, 0, 0)

		receipt = receipt + receipt1

		//loops through cloudcoins and starts placing information
		for e := range receiptinfo.CloudCoinCount {

			for i := 0; receiptinfo.CloudCoinCount[e] == i; i++ {
				fmt.Println(e, receiptinfo.CloudCoinCount[e])
				fmt.Println("index: ", e, i)
				ccCount := receiptinfo.CloudCoinCount[e]
				fmt.Println("\n", receipt)
				fmt.Println("Coincount in struct + ?: ", receiptinfo.CloudCoinCount[i], ccCount)

				//start of receipt
				if i == 0 {
					receiptDetail := fmt.Sprintf(
						`
					"denom": "%s",
					"nn.sn": "%s.%s",
					"status": "%s",
					"pown": "uuuuuuuuuuuuuuuuuuuuuuuuu",
					"note": "%s"
				}, `, receiptinfo.Denom[i], stack.CloudCoin[i].NN, stack.CloudCoin[i].SN, receiptinfo.Status[i], receiptinfo.Note[i])
					receipt = receipt + receiptDetail

					//middle of receipt
				} else if i > 0 && i < ccCount {

					receiptDetail := fmt.Sprintf(
						`
				{
					"denom": "%s",
					"nn.sn": "%s.%s",
					"status": "%s",
					"pown": "uuuuuuuuuuuuuuuuuuuuuuuuu",
					"note": "%s"
				},
				`, receiptinfo.Denom[i], stack.CloudCoin[i].NN, stack.CloudCoin[i].SN, receiptinfo.Status[i], receiptinfo.Note[i])
					receipt = receipt + receiptDetail

					//end of receipt
				} else if ccCount == 1 || i == ccCount {
					receiptDetail := fmt.Sprintf(
						`{
					"denom": "%s",
					"nn.sn": "%s.%s",
					"status": "%s",
					"pown": "uuuuuuuuuuuuuuuuuuuuuuuuu",
					"note": "%s"
				`, receiptinfo.Denom[i], stack.CloudCoin[i].NN, stack.CloudCoin[i].SN, receiptinfo.Status[i], receiptinfo.Note[i])
					receipt = receipt + receiptDetail
				} else {
					ErrStop(20, fmt.Errorf("error was made with index on receipt checker, index- %d cccount- %d", i, ccCount), t)
				} // end of else
			} //end of looping through invidual file
		} //end of looping through cc files

		receiptDetail := fmt.Sprint(
			`
			}

		]

	}`)
		//concatinates all parts
		receipt = receipt + receiptDetail

		receiptFolder := rootpath + `\logs\unpacker\receipt\`

		var _, err = os.Stat(receiptFolder)

		// create file if not exists
		if os.IsNotExist(err) {
			var file, err = os.Create(receiptFolder)
			if err != nil {
				ErrStop(21, err, t)
			}
			defer file.Close()
		} //end of if

		receiptFolder = receiptFolder + batchNumber + `.receipt` + `.txt`
		fmt.Println(receiptFolder)
		ioutil.WriteFile(receiptFolder, []byte(receipt), 0644)
	} //end of if

} //end of the func

//Denomination determines the denomination of the coin
func Denomination(sn int) int {
	var returnInt int
	returnInt = 0

	if sn >= 1 && sn <= 2097152 {
		returnInt = 1
	} else if sn <= 4194304 {
		returnInt = 5
	} else if sn <= 6291456 {
		returnInt = 25
	} else if sn <= 14680064 {
		returnInt = 100
	} else if sn <= 16777216 {
		returnInt = 250
	} else {
		ErrStop(28, fmt.Errorf("Denomination not recognized"), t)
	}
	return returnInt

} //end func denomination

//TrashedNumGenerator gives trashed coins a random number
func TrashedNumGenerator() string {
	rand.Seed(time.Now().UnixNano())
	num := strconv.Itoa(rand.Intn(999999))
	return num
} //end of TrashedNumGenerator

//GenGUI Makes a GUI for the receipt to use.
func GenGUI() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	uuid := fmt.Sprintf("%x%x%x%x%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid
}

//Fail exits the program with only a error number
func Fail(i int) {
	fmt.Printf("{\"status\":\"fail\",\"message\":\"error %d\"}", i)
	os.Exit(i)
}

//ErrStop checks if there is an error, and if so stops the program pringing the error message and the error number
func ErrStop(i int, err error, t time.Time) {
	if err != nil {
		errlog := fmt.Sprintf("{\"status\":\"fail\",\"message\":\"error %d. %s.  %v\"}", i, fmt.Sprintf("%s", err), time.Since(t))
		LogCoinMove(errlog, "No SN,", 3)
		os.Exit(i)
	} else {
		errlog := fmt.Sprintf("{\"status\":\"fail\",\"message\":\"error %d. %s.  %v\"}", i, fmt.Sprintf("Error that was not found by program. Could be the file is neither a directory or a regular."), time.Since(t))
		LogCoinMove(errlog, "No SN,", 3)
		os.Exit(i)
	}

}
