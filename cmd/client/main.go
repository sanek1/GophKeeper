package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/sanek1/GophKeeper/internal/client"
	"github.com/sanek1/GophKeeper/internal/models"
)

var (
	version    = "dev"
	commitHash = "none"
	buildDate  = ""
	buildNum   = "0"
)

func main() {
	cfg := client.Config{
		ServerURL:    "http://localhost:8080",
		SyncInterval: 5 * time.Minute,
	}

	c := client.NewClient(cfg)

	// Create a context for managing synchronization
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// run auto sync if user is authenticated
	if c.IsAuthenticated() {
		c.StartAutoSync(ctx)
		fmt.Println("Automatic synchronization started")
	}

	// handle signals for graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalCh
		fmt.Println("\nShutting down...")
		cancel() // cancel context to stop sync
		os.Exit(0)
	}()

	// check command line arguments
	if len(os.Args) > 1 {
		// if there are arguments, execute command and exit
		processCommand(c, ctx, os.Args[1:])
		return
	}

	// interactive mode
	fmt.Println("GophKeeper - manager of passwords and confidential data")
	fmt.Println("Version:", version)
	fmt.Println("Enter 'help' for a list of commands or 'exit' to exit")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		args := parseCommandLine(input)

		if len(args) == 0 {
			continue
		}

		if args[0] == "exit" || args[0] == "quit" {
			fmt.Println("Shutting down...")
			break
		}

		// Process the command
		processCommand(c, ctx, args)
	}
}

// parseCommandLine parses the input string into arguments
func parseCommandLine(input string) []string {
	input = strings.TrimSpace(input)
	if input == "" {
		return []string{}
	}
	return strings.Fields(input)
}

// processCommand processes commands
func processCommand(c *client.Client, ctx context.Context, args []string) {
	if len(args) == 0 {
		return
	}

	switch args[0] {
	case "help", "--help", "-h":
		printUsage()
		return

	case "version", "-v", "--version":
		printVersion()
		return

	case "register":
		if len(args) != 3 {
			fmt.Println("Usage: register <login> <password>")
			return
		}
		if err := c.Register(args[1], args[2]); err != nil {
			fmt.Printf("Registration error: %v\n", err)
			return
		}
		fmt.Println("Registration successful. Now you can log in.")

	case "login":
		if len(args) != 3 {
			fmt.Println("Usage: login <login> <password>")
			return
		}
		if err := c.Login(args[1], args[2]); err != nil {
			fmt.Printf("Login error: %v\n", err)
			return
		}

		// check token validity
		if err := c.TestAuthentication(); err != nil {
			fmt.Printf("Login successful, but token is invalid: %v\n", err)
			fmt.Println("Maybe the server is using a non-standard JWT_SECRET.")
			fmt.Println("Please check the server settings or try to log in again.")
			return
		}

		fmt.Println("Login successful.")
		// Start automatic synchronization after login
		c.StartAutoSync(ctx)

	case "list":
		// check if user is authenticated
		if !c.IsAuthenticated() {
			fmt.Println("You are not authenticated. Please log in.")
			return
		}

		// check token validity
		if err := c.TestAuthentication(); err != nil {
			fmt.Printf("Authentication error: %v\n", err)
			fmt.Println("Please log in again.")
			return
		}

		// check if master password is set
		if !c.IsMasterPasswordSet() {
			fmt.Println("Master password is not set. Use the set-master-password command.")
			return
		}

		secrets, err := c.GetSecrets()
		if err != nil {
			// if there is no connection to the server, use local data
			fmt.Printf("Warning: %v\n", err)
			fmt.Println("Using local data from cache...")
			secrets = c.GetOfflineSecrets()
		}

		if len(secrets) == 0 {
			fmt.Println("Secrets not found")
			return
		}

		fmt.Println("Your secrets:")
		for _, s := range secrets {
			fmt.Printf("ID: %s, Type: %s, Metadata: %s\n", s.ID, s.Type, s.Metadata)
		}

	case "create":
		if len(args) != 4 {
			fmt.Println("Usage: create <type> <metadata> <data>")
			fmt.Println("Supported types:", models.SecretTypes)
			return
		}

		// check if user is authenticated
		if !c.IsAuthenticated() {
			fmt.Println("You are not authenticated. Please log in.")
			return
		}

		// check token validity
		if err := c.TestAuthentication(); err != nil {
			fmt.Printf("Authentication error: %v\n", err)
			fmt.Println("Please log in again.")
			return
		}

		// check if master password is set
		if !c.IsMasterPasswordSet() {
			fmt.Println("Master password is not set. Use the set-master-password command.")
			return
		}

		secretType := args[1]
		metadata := args[2]
		data := []byte(args[3])

		// check if secret type is valid
		validType := false
		for _, t := range models.SecretTypes {
			if t == secretType {
				validType = true
				break
			}
		}
		if !validType {
			fmt.Printf("Invalid secret type. Supported types: %v\n", models.SecretTypes)
			return
		}

		secret, err := c.CreateSecret(secretType, metadata, data)
		if err != nil {
			fmt.Printf("Error creating secret: %v\n", err)
			return
		}
		fmt.Printf("Secret created with ID: %s\n", secret.ID)

	case "get":
		if len(args) != 2 {
			fmt.Println("Usage: get <id>")
			return
		}

		// check if user is authenticated
		if !c.IsAuthenticated() {
			fmt.Println("You are not authenticated. Please log in.")
			return
		}

		// check token validity
		if err := c.TestAuthentication(); err != nil {
			fmt.Printf("Authentication error: %v\n", err)
			fmt.Println("Please log in again.")
			return
		}

		// check if master password is set
		if !c.IsMasterPasswordSet() {
			fmt.Println("Master password is not set. Use the set-master-password command.")
			return
		}

		secret, err := c.GetSecret(args[1])
		if err != nil {
			fmt.Printf("Error getting secret: %v\n", err)
			return
		}

		fmt.Printf("ID: %s\nType: %s\nMetadata: %s\n",
			secret.ID, secret.Type, secret.Metadata)

		// use new function to display data
		data, err := c.DisplaySecretData(secret)
		if err != nil {
			fmt.Printf("Error displaying data: %v\n", err)
			return
		}
		fmt.Printf("Data: %s\n", data)

	case "delete":
		if len(args) != 2 {
			fmt.Println("Usage: delete <id>")
			return
		}

		// check if user is authenticated
		if !c.IsAuthenticated() {
			fmt.Println("You are not authenticated. Please log in.")
			return
		}

		// check token validity
		if err := c.TestAuthentication(); err != nil {
			fmt.Printf("Authentication error: %v\n", err)
			fmt.Println("Please log in again.")
			return
		}

		if err := c.DeleteSecret(args[1]); err != nil {
			fmt.Printf("Error deleting secret: %v\n", err)
			return
		}
		fmt.Println("Secret successfully deleted")

	case "update":
		if len(args) != 4 {
			fmt.Println("Usage: update <id> <metadata> <data>")
			return
		}

		// check if user is authenticated
		if !c.IsAuthenticated() {
			fmt.Println("You are not authenticated. Please log in.")
			return
		}

		// check token validity
		if err := c.TestAuthentication(); err != nil {
			fmt.Printf("Authentication error: %v\n", err)
			fmt.Println("Please log in again.")
			return
		}

		// check if master password is set
		if !c.IsMasterPasswordSet() {
			fmt.Println("Master password is not set. Use the set-master-password command.")
			return
		}

		id := args[1]
		metadata := args[2]
		data := []byte(args[3])

		// get current secret to save its type
		secret, err := c.GetSecret(id)
		if err != nil {
			fmt.Printf("Error getting secret: %v\n", err)
			return
		}

		if err := c.UpdateSecret(id, secret.Type, metadata, data); err != nil {
			fmt.Printf("Error updating secret: %v\n", err)
			return
		}
		fmt.Println("Secret successfully updated")

	case "set-master-password":
		if len(args) != 2 {
			fmt.Println("Usage: set-master-password <password>")
			return
		}
		if err := c.SetMasterPassword(args[1]); err != nil {
			fmt.Printf("Error setting master password: %v\n", err)
			return
		}
		fmt.Println("Master password successfully set")

	case "sync":
		// command for manual synchronization
		if !c.IsAuthenticated() {
			fmt.Println("You are not authenticated. Please log in.")
			return
		}

		// check token validity
		if err := c.TestAuthentication(); err != nil {
			fmt.Printf("Authentication error: %v\n", err)
			fmt.Println("Please log in again.")
			return
		}

		fmt.Println("Synchronizing data with the server...")
		if err := c.SyncWithServer(); err != nil {
			fmt.Printf("Synchronization error: %v\n", err)
			return
		}
		fmt.Println("Synchronization completed successfully")

	case "logout":
		// command for logout
		if err := c.Logout(); err != nil {
			fmt.Printf("Logout error: %v\n", err)
			return
		}
		fmt.Println("You have logged out")

	case "clear", "cls":
		// clear screen
		clearScreen()

	default:
		fmt.Printf("Unknown command: %s\nEnter 'help' for a list of commands\n", args[0])
	}
}

func printUsage() {
	fmt.Println("GophKeeper - manager of passwords and confidential data")
	fmt.Println("Version:", version)
	fmt.Println("\nAvailable commands:")
	fmt.Println("  help - show this list of commands")
	fmt.Println("  version - show version information")
	fmt.Println("  register <login> <password> - register a new user")
	fmt.Println("  login <login> <password> - log in to the system")
	fmt.Println("  set-master-password <password> - set a master password for encryption")
	fmt.Println("  list - show the list of secrets")
	fmt.Println("  create <type> <metadata> <data> - create a new secret")
	fmt.Println("  get <id> - get a secret by ID")
	fmt.Println("  update <id> <metadata> <data> - update a secret")
	fmt.Println("  delete <id> - delete a secret by ID")
	fmt.Println("  sync - synchronize data with the server")
	fmt.Println("  logout - log out of the system")
	fmt.Println("  clear - clear the screen")
	fmt.Println("  exit - exit the program")
	fmt.Println("\nSupported secret types:", models.SecretTypes)
}

func printVersion() {
	info := models.ClientInfo{
		Version:     version,
		BuildDate:   buildDate,
		CommitHash:  commitHash,
		BuildNumber: buildNum,
	}

	fmt.Println("GophKeeper Client")
	fmt.Println("Version:", info.Version)
	fmt.Println("Build:", info.BuildNumber)
	fmt.Println("Build date:", info.BuildDate)
	fmt.Println("Commit:", info.CommitHash)
	fmt.Println("OS/Architecture:", runtime.GOOS+"/"+runtime.GOARCH)
}

// clearScreen clears the screen depending on the platform
func clearScreen() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		// For Unix-like systems (Linux, macOS)
		fmt.Print("\033[H\033[2J")
	}
}
