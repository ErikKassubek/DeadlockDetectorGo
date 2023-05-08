package main

/*
Copyright (c) 2023, Erik Kassubek
All rights reserved.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

/*
Author: Erik Kassubek <erik-kassubek@t-online.de>
Package: dedego-instrumenter
Project: Dynamic Analysis to detect potential deadlocks in concurrent Go programs
*/

/*
main.go
main file to run the program
*/

import (
	"errors"
	"fmt"
	"math/rand"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/ErikKassubek/deadlockDetectorGo/src/gui"
	"github.com/ErikKassubek/deadlockDetectorGo/src/instrumenter"

	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

/*
Function to run the program
@param elements *gui.GuiElements: gui elements to display output
@param status *gui.Status: status of the program
@return error: error or nil
*/
func run(elements *gui.GuiElements, status *gui.Status) error {
	if status.Instrument {
		elements.AddToOutput("Starting Instrumentation")
	} else {
		elements.AddToOutput("Analyse files")
	}

	// check numeric elements
	r, _ := regexp.Compile("^[0-9][0-9]*$")
	r0, _ := regexp.Compile("^[1-9][0-9]*$")
	r1 := r0.MatchString(elements.SettingsMaxRuns.Text)
	r3 := r.MatchString(elements.SettingMaxTime.Text)
	r4 := r.MatchString(elements.SettingMaxSelectTime.Text)
	if !r1 {
		elements.AddToOutput("Max Runs must be numeric and greater or equal to 1")
		return fmt.Errorf("max Runs and Max Failed must be numeric")
	}
	if !r3 || !r4 {
		elements.AddToOutput("Times must be numeric (int)")
		return fmt.Errorf("max times must be numeric")
	}
	status.SettingsMaxRuns, _ = strconv.Atoi(elements.SettingsMaxRuns.Text)
	status.SettingsMaxFailed = status.SettingsMaxRuns / 2
	status.SettingMaxTime, _ = strconv.Atoi(elements.SettingMaxTime.Text)
	status.SettingMaxSelectTime, _ = strconv.Atoi(
		elements.SettingMaxSelectTime.Text)
	if status.SettingMaxSelectTime <= 0 {
		status.SettingMaxSelectTime = 2147483647
	}

	// instrument files
	select_map, err := instrumenter.InstrumentFiles(elements, status)
	if err != nil {
		if status.Instrument {
			elements.AddToOutput("Instrumentation Failed: " + err.Error())
		} else {
			elements.AddToOutput("Analysis Failed: " + err.Error())
		}
		return err
	} else {
		if status.Instrument {
			elements.AddToOutput("Instrumentation Complete\n")
		} else {
			elements.AddToOutput("Analysis Complete\n")
		}
	}

	if !status.Instrument {
		status.Output = status.FolderPath
	}

	// install analyzer
	elements.AddToOutput("Installing Analyzer")
	elements.ProgressBuild.SetValue(0.1)
	cmd := exec.Command("go", "get",
		"github.com/ErikKassubek/deadlockDetectorGo/src/dedego")
	if status.Instrument {
		cmd.Dir = status.Output + string(os.PathSeparator) + status.Name
	} else {
		cmd.Dir = status.Output
	}
	out, err := cmd.Output()
	if len(out) > 0 {
		elements.AddToOutput(string(out) + "")
	}
	if err != nil {
		elements.AddToOutput("Failed to install Analyzer: " + err.Error())
		return err
	}
	elements.AddToOutput("Analyzer installed\n")
	elements.ProgressBuild.SetValue(0.2)

	// cleanup files
	if status.Instrument {
		elements.AddToOutput("Cleaning up files")
		cmd = exec.Command("goimports", "-w", ".")
		cmd.Dir = status.Output + string(os.PathSeparator) + status.Name
		out, err = cmd.Output()
		if len(out) > 0 {
			elements.AddToOutput(string(out))
		}
		if err != nil {
			elements.AddToOutput("Failed to cleanup files: " + err.Error())
			return err
		}
		elements.ProgressBuild.SetValue(0.3)
		elements.AddToOutput("Files cleaned up\n")
	}

	elements.AddToOutput("Building program")
	cmd = exec.Command("go", "build", "-o", "dedego_instrumented")
	if status.Instrument {
		cmd.Dir = status.Output + string(os.PathSeparator) + status.Name
	} else {
		cmd.Dir = status.Output
	}
	out, err = cmd.Output()
	if len(out) > 0 {
		elements.AddToOutput(string(out))
	}
	if err != nil {
		elements.AddToOutput("Failed to build program: " + err.Error())
		return err
	}
	elements.ProgressBuild.SetValue(1)
	elements.AddToOutput("Program built\n")

	// analyse the program
	res := analyse(select_map, status, elements)

	return res
}

/*
Get a string from a switch order
@param soe map[int]int: order
@return string: string representing the switch order
*/
func toString(soe map[int]int) string {
	res := ""
	i := 0
	for key, c := range soe {
		res += fmt.Sprint(key) + "," + fmt.Sprint(c)
		if i != len(soe)-1 {
			res += ";"
		}
		i++
	}
	return res
}

/*
Check if an order was not inserted into the queue before
@param order map[int]int: map representing an order
@return bool: true, if the order was not in the queue before, false otherwise
*/
func wasNotInQueue(order map[int]int, queue *[]map[int]int) bool {
	for _, i := range *queue {
		if reflect.DeepEqual(i, order) {
			return false
		}
	}
	return true
}

/*
Check if elem is in list
@param list *[]string: list to check
@param elem string: element to check
@return bool: true if elem is in list, false otherwise
*/
func isInList(list []string, elem string) bool {
	for _, e := range list {
		if e == elem {
			return true
		}
	}
	return false
}

/*
Run the analysis on the instrumented program
@param switch_size map[int]int: map of switch sizes
@param status *gui.Status: status of the program
@param elements *gui.GuiElements: gui elements to display output
@return error: error or nil
*/
func analyse(switch_size map[int]int, status *gui.Status,
	elements *gui.GuiElements) error {
	// initialize
	rand.Seed(time.Now().UTC().UnixNano())
	var queue = make([]map[int]int, 0)       // orders to test
	var messages = make(map[string][]string) // collect messages
	failed := false

	elements.AddToOutput("Starting Analysis")
	elements.ProgressAna.SetValue(0.02)

	elements.AddToOutput("Determine switch execution order")
	elements.ProgressAna.SetValue(0.05)

	no_failed := status.SettingsMaxFailed
	max_runs := status.SettingsMaxRuns
	for no_failed > 0 && max_runs > 0 {
		order_add := make(map[int]int)
		for key, size := range switch_size {
			if size <= 0 {
				size = 1
			}
			order_add[key] = rand.Intn(size)
		}
		if wasNotInQueue(order_add, &queue) {
			queue = append(queue, order_add)
		} else {
			no_failed -= 1
		}
		max_runs -= 1
	}

	elements.AddToOutput("Starting Program Execution\n")
	elements.ProgressAna.SetValue(0.1)

	for i := 0; i < len(queue); i++ {
		order := queue[i]
		orderString := toString(order)

		var cmd *exec.Cmd

		if status.Instrument {
			os.Chdir(os.TempDir() + string(os.PathSeparator) + "dedego" +
				string(os.PathSeparator) + status.Name)
		} else {
			os.Chdir(status.Output)
		}
		command := "./" + "dedego_instrumented"

		print(command, orderString)

		if orderString == "" {
			cmd = exec.Command(command)
		} else {
			cmd = exec.Command(command, orderString)
		}
		out, err := cmd.CombinedOutput()
		if err != nil {
			if fmt.Sprint(err) == "exit status 42" {
				elements.AddToOutput("Runtime exceeded limit")
			} else {
				failed = true
				elements.AddToOutput(err.Error())
			}
		}

		output := string(out)

		start := false
		if !strings.HasPrefix(output, "##@@##") {
			start = true
		}

		for i, message := range strings.Split(output, "##@@##") {
			if len(strings.TrimSpace(message)) == 0 {
				continue
			}
			if strings.Contains(message, "panic: send on closed channel") {
				m_split := strings.Split(message, "\n")
				message = "Send on closed channel:\n    " +
					m_split[len(m_split)-2]
			}
			if start && i == 0 {
				continue
			}
			if _, ok := messages[message]; !ok {
				messages[message] = make([]string, 0)
			}

			if !isInList(messages[message], orderString) {
				messages[message] = append(messages[message], orderString)
			}
		}
		elements.ProgressAna.SetValue(0.1 + (0.9 *
			float64(i+1) / float64(len(queue))))
	}

	l := len(messages)
	if l == 0 && !failed {
		elements.AddToOutput("No Problems Found\n")
	} else if l > 0 {
		elements.AddToOutput("Found Problems:\n")
		for message, orders := range messages {
			if len(orders) != 0 && len(message) != 0 &&
				strings.TrimSpace(orders[0]) != "" {
				elements.AddToOutput(
					"Found while examine the following orders: ")
				for _, order := range orders {
					elements.AddToOutput("  " + order)
				}
				elements.AddToOutput("\n")
			}
			elements.AddToOutput(message)
			elements.AddToOutput("\n")
		}
		elements.AddToOutput(
			"Note: The positions show the positions in the instrumented code!")
	}
	if failed {
		return errors.New("analysis failed")
	}
	elements.ProgressAna.SetValue(1)
	return nil
}

/*
Main function
*/
func main() {
	app := app.New()
	window := app.NewWindow("Deadlock Detector Go")
	status := gui.Status{}
	status.Output = os.TempDir() + string(os.PathSeparator) + "dedego"
	elements := gui.GuiElements{}

	// create elements
	elements.PathLab = widget.NewLabel("Path:")
	elements.Output = widget.NewTextGrid()
	elements.OutputScroll = container.NewScroll(elements.Output)

	// create a scroll container for the output
	elements.OpenBut = widget.NewButton("Open", nil)
	elements.StartBut = widget.NewButton("Start", nil)

	// create settings
	elements.SettingsMaxRuns = widget.NewEntry()
	elements.SettingMaxTime = widget.NewEntry()
	elements.SettingMaxSelectTime = widget.NewEntry()
	elements.InfoText = widget.NewLabel("Those settings are only used if the input code is not instrumented!")
	buttonText := "Input code is instrumented"
	if status.Instrument {
		buttonText = "Input code is not instrumented"
	}
	elements.InstrumentButton = widget.NewButton(buttonText, func() {
		status.Instrument = !status.Instrument
		if status.Instrument {
			elements.InstrumentButton.SetText("Input code is instrumented")
		} else {
			elements.InstrumentButton.SetText("Input code is not instrumented")
		}
	})
	elements.Settings = widget.NewForm(
		widget.NewFormItem("Max number of runs", elements.SettingsMaxRuns),
		widget.NewFormItem("Max wait time per run (s)",
			elements.SettingMaxTime),
		widget.NewFormItem("Max wait time per select (s)",
			elements.SettingMaxSelectTime),
	)
	elements.SettingsMaxRuns.SetText("20")
	elements.SettingMaxTime.SetText("20")
	elements.SettingMaxSelectTime.SetText("2")

	// create progress bars
	elements.ProgressInst = widget.NewProgressBar()
	elements.ProgressBuild = widget.NewProgressBar()
	elements.ProgressAna = widget.NewProgressBar()
	elements.Progress = widget.NewForm(
		widget.NewFormItem("Instrumentation", elements.ProgressInst),
		widget.NewFormItem("Build", elements.ProgressBuild),
		widget.NewFormItem("Analysis", elements.ProgressAna),
	)

	// set button functions

	elements.OpenBut.OnTapped = func() {
		// BUG: cancel creates segmentation violation
		fileDialog := dialog.NewFolderOpen(
			func(r fyne.ListableURI, _ error) {
				status.FolderPath = r.Path()
				splitPath := strings.Split(status.FolderPath,
					string(os.PathSeparator))
				status.Name = splitPath[len(splitPath)-1]
				elements.PathLab.SetText("Path: " + status.FolderPath)
			}, window)
		fileDialog.Show()
	}

	elements.StartBut.OnTapped = func() {
		if status.FolderPath == "" {
			elements.Output.SetText("No folder selected!")
			return
		} else {
			elements.ProgressBuild.SetValue(0)
			elements.ProgressAna.SetValue(0)
			elements.ProgressAna.SetValue(0)
			elements.Progress.Hidden = false
			elements.ClearOutput()
			go func() {
				err := run(&elements, &status)
				elements.AddToOutput("")
				if err != nil {
					elements.AddToOutput("Analysis failed")
				} else {
					elements.AddToOutput("Analysis complete")
				}
			}()
		}
	}

	// set initial state
	elements.Progress.Hidden = true

	// create layout
	gridUp := container.NewVBox(elements.PathLab, elements.OpenBut,
		elements.InstrumentButton,
		elements.Settings, elements.InfoText,
		elements.StartBut, elements.Progress)
	grid := container.New(layout.NewGridLayout(1), gridUp,
		elements.OutputScroll)

	// show window
	window.SetContent(grid)
	window.ShowAndRun()
}
