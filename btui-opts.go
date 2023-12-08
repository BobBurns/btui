package main

import (
	"github.com/pterm/pterm"
	"log"
	"os/exec"
	"os/user"
	"io"
	"io/ioutil"
	"os"
	"encoding/json"

)
type Server struct {
        Ipaddr     string `json:"ipaddr"`
	Iface  string `json:"iface"`
        Available string `json:"available"`
        Running string `json:"running"`
        Name string `json:"name"`
        Owner string `json:"owner"`
        Type string `json:"type"`
}

type VM struct {
	Servers []Server `json:"servers"`
}

func printHeader(h string) {
	print("\033[H\033[2J")
	header := pterm.DefaultHeader
	header.BackgroundStyle = pterm.NewStyle(pterm.BgBlue)
	header.WithFullWidth().Println(h)
	pterm.Println() // Blank line
}
func goodbye() {
	pterm.Println()
	e := pterm.Success
	e.Prefix.Text = "GoodBye!"
	e.Printf("Please Log out of the system\n")
	os.Exit(0)
}

func checkerr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func continueText() {
	confirm := pterm.DefaultInteractiveContinue
	confirm.DefaultText = "press <enter> to continue"
	confirm.Options = []string{"ok"}
	result, _ := confirm.Show()
	pterm.Println(result)
}

func showContainers(vms VM) {
	ptable := [][]string{}

//	pterm.DefaultHeader.WithFullWidth().Println("Available containers:")
	pterm.Println(pterm.BgBlue.Sprint("Available containers:"))
	pterm.Println(pterm.Bold.Sprint("============================="))

	// list json table
	ptable = append(ptable, []string{"IP Address", "Name", "Owner", "Running", "Available", "Type"})
	for _,v := range vms.Servers {
		ptable = append(ptable, []string{v.Ipaddr, v.Name, v.Owner, v.Running, v.Available, v.Type})
	}
	pterm.DefaultTable.WithHasHeader().WithData(ptable).Render()
}

func serverAction(vms VM, a string) {
	h := pterm.Sprintf("%s Server\n", a)
	printHeader(h)
	showContainers(vms)
	servers := []string{}
	cuser, err := user.Current()
	checkerr(err)
	
	for _, v := range vms.Servers {
		if v.Available == "no" && v.Owner == cuser.Username {
			servers = append(servers, v.Name)
		}
	}
	if len(servers) < 1 {
		pterm.Println("no servers for", cuser.Username)
		continueText()
		return
	}
	newselect := pterm.DefaultInteractiveSelect
	newselect.MaxHeight = 7
	newselect.DefaultText =  pterm.Sprintf("Please select server for %s to %s", pterm.Green(cuser.Username), a)
	
	servOpt, _ := newselect.WithOptions(servers).Show()
	pterm.Info.Printfln("Do you want to %s %s? ", a, servOpt)
	confirmB, _ := pterm.DefaultInteractiveConfirm.Show()
	if !confirmB {
		return
	}

	for i, v := range vms.Servers {
		if v.Name == servOpt {
			switch a {
			case "start":
				if v.Running == "yes" {
					pterm.Warning.Println("Server is already running.")
					continueText()
					return
				}
				vms.Servers[i].Running = "yes"
			case "stop":
				if v.Running == "no" {
					pterm.Warning.Println("Server is not running.")
					continueText()
					return
				}
				vms.Servers[i].Running = "no"
			case "destroy":
				if v.Running == "yes" {
					pterm.Warning.Println("Server must be stopped before destroy")
					continueText()
					return
				}
				vms.Servers[i].Name = "none"
				vms.Servers[i].Available = "yes"
				vms.Servers[i].Type = "none"
				vms.Servers[i].Owner = "none"
				
			}
		}
	}
	cmd := exec.Command("doas", "/root/bastille-scripts/action.csh", servOpt, a)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	slurp, _ := io.ReadAll(stderr)
	pterm.Info.Printf("%s\n", slurp)

	if err := cmd.Wait(); err != nil {
		pterm.Fatal.Println("Something went wrong...")
		log.Fatal(err)
	}
	
	
	jsonOut, err := json.MarshalIndent(vms,"  ","  ") 
	checkerr(err)

  	ioutil.WriteFile("/usr/local/etc/vms.json", jsonOut, 0666)
	continueText()
}

func bastilleList() {

	cmd := exec.Command("doas", "/root/bastille-scripts/action.csh", "all", "list")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	slurp, _ := io.ReadAll(stdout)
	pterm.Info.Printf("%s\n", slurp)

	if err := cmd.Wait(); err != nil {
		pterm.Fatal.Println("Something went wrong...")
		log.Fatal(err)
	}
	
	
	continueText()
}

func buildServer(vms VM) {
	printHeader("Build Server")
	showContainers(vms)
	pterm.DefaultBasicText.Println("Please choose a server type to build")


	// select options
	var options []string
	options = append(options, "Golang Server")
	options = append(options, "R Server")
	options = append(options, "Nginx Server")
	options = append(options, "Python Dev")
	options = append(options, "Rust Dev")
	options = append(options, "Ruby Dev")
	options = append(options, "Node Dev")
	options = append(options, "Apache Server")
	options = append(options, "Iperf Server")

	selectedOption, _ := pterm.DefaultInteractiveSelect.WithOptions(options).Show()
	pterm.Info.Printfln("Selected option: %s", pterm.Green(selectedOption))
	sOpt := "none"
	switch selectedOption {
	case "Golang Server":
		sOpt = "gdev"
	case "R Server":
		sOpt = "R-stat"
	case "Nginx Server":
		sOpt = "nginx"
	case "Python Dev":
		sOpt = "python"
	case "Rust Dev":
		sOpt = "rust"
	case "Ruby Dev":
		sOpt = "ruby"
	case "Node Dev":
		sOpt = "node"
	case "Apache Server":
		sOpt = "apache"
	case "Iperf Server":
		sOpt = "iperf"
	}
	
	// get ssh key
	pterm.Println() // Blank line
	pterm.DefaultBasicText.Println("Please paste in your ssh pub key")
	result, _ := pterm.DefaultInteractiveTextInput.WithMultiLine(false).Show()
	pterm.Println() // Blank line
	pterm.Info.Printfln("You answered: %s", result)
	result = result + "\n"

  	ioutil.WriteFile("/tmp/authkey", []byte(result), 0600)
	// get name
	pterm.Println() // Blank line
	pterm.DefaultBasicText.Println("Enter name of your container")
	cName, _ := pterm.DefaultInteractiveTextInput.WithMultiLine(false).Show()
	pterm.Println() // Blank line
	pterm.Info.Printfln("Continue building %s? \n", cName)
	confirmB, _ := pterm.DefaultInteractiveConfirm.Show()
	if !confirmB {
		return
	}

	spinnerInfo, _ := pterm.DefaultSpinner.Start("Starting build. Please wait...")
	cuser, err := user.Current()
	checkerr(err)

	// get available server
	availContainer := Server{}
	for i,v := range vms.Servers {
		if v.Available == "yes" {
			availContainer = Server{
				Ipaddr: v.Ipaddr,
				Name: cName,
				Iface: v.Iface,
				Running: "yes",
				Available: "no",
				Owner: cuser.Username,
				Type: sOpt,
			}
			vms.Servers[i] = availContainer
			break
		}
	}

	// cmd example
	// echo usage create.csh <jid> <int> <ipaddr> <type>
	cmd := exec.Command("doas", "/root/bastille-scripts/create-jail.csh", availContainer.Name, availContainer.Iface, availContainer.Ipaddr, availContainer.Type)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	slurp, _ := io.ReadAll(stderr)
	pterm.Info.Printf("%s\n", slurp)

	if err := cmd.Wait(); err != nil {
		spinnerInfo.Fail("Something went wrong")
		log.Fatal(err)
	}

	
	jsonOut, err := json.MarshalIndent(vms,"  ","  ") 
	checkerr(err)

  	ioutil.WriteFile("/usr/local/etc/vms.json", jsonOut, 0666)
	spinnerInfo.Success("Congratulations! You can now log in to your new servers as root to ", availContainer.Ipaddr)
	continueText()
}

func main() {

	for {
		file, err := os.Open("/usr/local/etc/vms.json")
		checkerr(err)
		defer file.Close()

		var vms VM

		b, err := io.ReadAll(file)
		checkerr(err)
		err = json.Unmarshal(b, &vms)
		checkerr(err)

		
		printHeader("Welcome to Bastille interactive container tool")

		// select action
		messText := pterm.DefaultBasicText
		messText.Println("Build and manage your own FreeBSD container from software templates.")


		// select options
		var actopts []string
		actopts = append(actopts, "Build Server")
		actopts = append(actopts, "Start Server")
		actopts = append(actopts, "Stop Server")
		actopts = append(actopts, "Destroy Server")
		actopts = append(actopts, "List Servers")
		actopts = append(actopts, "Bastille List Servers")
		actopts = append(actopts, "Quit")

		myselect := pterm.DefaultInteractiveSelect
		myselect.MaxHeight = 7
		selActopt, _ := myselect.WithOptions(actopts).Show()

		switch selActopt {
		case "Build Server":
			buildServer(vms)
		case "Start Server":
			serverAction(vms, "start")
		case "Stop Server":
			serverAction(vms, "stop")
		case "Destroy Server":
			serverAction(vms, "destroy")
		case "Bastille List Servers":
			bastilleList()
		case "List Servers":
			showContainers(vms)
			continueText()
		case "Quit":
			goodbye()
		}
	}
}
