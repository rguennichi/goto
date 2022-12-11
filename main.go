package main

import (
	"fmt"
	"github.com/desertbit/grumble"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
)

func main() {
	app := grumble.New(&grumble.Config{
		Name:                  "goto",
		Description:           "Quick SSH access to application servers/environments",
		PromptColor:           color.New(color.FgGreen).Add(color.Bold),
		HelpHeadlineColor:     color.New(color.FgYellow),
		HelpHeadlineUnderline: true,
		HelpSubCommands:       true,
		Flags: func(f *grumble.Flags) {
			ex, err := os.Executable()
			if err != nil {
				log.Fatal(err)
			}

			f.String("c", "config", filepath.Join(filepath.Dir(ex), "goto.yaml"), "Configuration file")
		},
	})

	app.SetPrintASCIILogo(func(a *grumble.App) {
		fmt.Println("             _           ")
		fmt.Println("  __ _  ___ | |_ ___     ")
		fmt.Println(" / _` |/ _ \\| __/ _ \\  ")
		fmt.Println("| (_| | (_) | || (_)     ")
		fmt.Println(" \\__, |\\___/ \\__\\___/")
		fmt.Println(" |___/                   ")
		fmt.Println("")
	})

	app.OnShell(func(gapp *grumble.App) error {
		// Ignore interrupt signals, because grumble will handle the interrupts anyway.
		// and the interrupt signals will be passed through automatically to all
		// client processes. They will exit, but the shell will pop up and stay alive.
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)
		go func() {
			for {
				<-signalChan
			}
		}()
		return nil
	})

	app.OnInit(func(a *grumble.App, flags grumble.FlagMap) error {
		// Parse YAML config file
		// And create commands based on the configuration.
		m, err := Parse(flags.String("config"))
		if err != nil {
			log.Fatal(err)
		}

		for appname, a := range m.Applications {
			for e, em := range a.Server.Environments {
				var (
					// Catch variables locally for run.
					env  = e
					envm = em
					appm = a
				)
				sshCmd := &grumble.Command{
					Name: appname + "@" + env,
					Help: "Application SSH",
					Run: func(c *grumble.Context) error {
						promptInstance := promptui.Select{
							Label: "Select instance",
							Items: envm.Hosts,
						}

						_, host, err := promptInstance.Run()
						if err != nil {
							return nil
						}

						if env == "prod" {
							promptConfirm := promptui.Prompt{
								Label:     "Continue to PRODUCTION",
								IsConfirm: true,
							}

							confirm, err := promptConfirm.Run()
							if err != nil {
								return nil
							}

							if confirm != "y" {
								fmt.Println("Canceled")
								return nil
							}
						}

						// Run SSH
						prefix := "set -x\n"
						sshCmd := fmt.Sprintf("ssh -l %s -p %s %s -t 'sudo -iu %s -- sh -c \"cd %s; /bin/bash\"; bash -'",
							appm.Server.Username,
							appm.Server.Port,
							host,
							appm.Username,
							appm.Path,
						)
						cmd := exec.Command("bash", "-c", prefix+sshCmd)

						cmd.Stdin = os.Stdin
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr

						return cmd.Run()
					},
				}

				app.AddCommand(sshCmd)

				execCmd := &grumble.Command{
					Name: "exec",
					Help: "Execute script",
					Run: func(c *grumble.Context) error {
						return nil
					},
				}

				sshCmd.AddCommand(execCmd)

				for _, s := range appm.Scripts {
					var (
						script = s
					)
					execCmd.AddCommand(&grumble.Command{
						Name: script.Name,
						Help: strOrDefault(script.Desc, "No description"),
						Run: func(c *grumble.Context) error {
							promptInstance := promptui.Select{
								Label: "Select instance",
								Items: envm.Hosts,
							}

							_, host, err := promptInstance.Run()
							if err != nil {
								return nil
							}

							if env == "prod" {
								promptConfirm := promptui.Prompt{
									Label:     "Execute \"" + script.Name + "\" on PRODUCTION",
									IsConfirm: true,
								}

								confirm, err := promptConfirm.Run()
								if err != nil {
									return nil
								}

								if confirm != "y" {
									fmt.Println("Canceled")
									return nil
								}
							}

							prefix := "set -x\n"
							sshCmd := fmt.Sprintf("ssh -l %s -p %s %s -t 'sudo -iu %s -- sh -c \"cd %s && %s; /bin/bash\"'",
								appm.Server.Username,
								appm.Server.Port,
								host,
								appm.Username,
								appm.Path,
								script.Exec,
							)

							cmd := exec.Command("bash", "-c", prefix+sshCmd)

							cmd.Stdin = os.Stdin
							cmd.Stdout = os.Stdout
							cmd.Stderr = os.Stderr

							return cmd.Run()
						},
					})
				}
			}
		}

		return nil
	})

	grumble.Main(app)
}

func strOrDefault(str string, defaultStr string) string {
	if str == "" {
		return defaultStr
	}

	return str
}
