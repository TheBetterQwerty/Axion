package main

import (
	"fmt"
	"axion/utils"
	"encoding/json"
	"net"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var inputChan = make(chan string);
var outputChan = make(chan string);
var chatList *tview.List;
var app *tview.Application;
var username string;

func main() {
	app = tview.NewApplication();
	var password string;

	input := tview.NewInputField().
		SetLabel("Server Ip").
		SetText("127.0.0.1:8080").
		SetFieldWidth(30);

	form := tview.NewForm().
		AddFormItem(input).
		AddInputField("Username", "", 30, nil, func(text string) {
			username = text;
		}).
		AddPasswordField("Password", "", 30, '*', func(text string) {
			password = text;
		}).
		AddButton("Submit!", func() {
			if strings.TrimSpace(username) == "" {
				return;
			}

			sockfd, err := net.Dial("tcp", input.GetText());
			if err != nil {
				panic(fmt.Sprintf("Failed to connect to server: %x\n", err));
			}

			pkt := axion.New(username, "SERVER");
			encoded, _ := json.Marshal(pkt);
			sockfd.Write(encoded);

			key := axion.GetKey(password); // should take password and return the hash

			go handle_server_read(sockfd, key);
			go handle_server_write(sockfd, key);

			ChatUI(username);
		}).
		AddButton("Quit", func() {
			app.Stop()
		})

	form.
		SetButtonsAlign(tview.AlignCenter).
		SetBorder(true).
		SetTitle(" [::b][red]Create Chatroom[::-] ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderPadding(1, 1, 2, 2)

	// Label above form
	desc := tview.NewTextView().
		SetText("> Connect to the chat server.").
		SetTextAlign(tview.AlignLeft).
		SetTextColor(tcell.ColorDarkGray)

	// Stack description and form vertically
	formFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(desc, 1, 1, false).
		AddItem(form, 0, 1, true)

	// Center the whole thing
	centered := tview.NewFlex().
		AddItem(nil, 0, 1, false).                 // left spacer
		AddItem(formFlex, 60, 0, true).            // form box width
		AddItem(nil, 0, 1, false)                  // right spacer
	finalLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).                 // top spacer
		AddItem(centered, 15, 0, true).            // form box height
		AddItem(nil, 0, 1, false);                  // bottom spacer

	if err := app.SetRoot(finalLayout, true).Run(); err != nil {
		panic(err)
	}
}

func ChatUI(username string) {
	chatList = tview.NewList().ShowSecondaryText(false);
	usr_color := axion.Get_color(username);

	chatList.SetBorder(true).SetTitle(" Users ");
	chatList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return nil;
	})

	// Message area
	messageView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			app.Draw()
		});
	messageView.SetBorder(true).SetTitle(" Messages ");

	// Input field at the bottom
	inputField := tview.NewInputField().
		SetLabel(fmt.Sprintf("%s: ", username)).
		SetFieldWidth(0);

	var formated_str string;
	inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			text := inputField.GetText()
			if strings.TrimSpace(text) != "" {
				if data, found := strings.CutPrefix(text, "/msg "); found {
					splits := strings.Split(data, " ");
					if len(splits) < 2 {
						return;
					}
					usr2_color := axion.Get_color(splits[0]);
					formated_str = fmt.Sprintf("[%s][ %s[-][%s] (%s) ][-] %s\n", usr_color, username,
						usr2_color, splits[0],
						strings.Join(splits[1:], " "));
				} else {
					formated_str = fmt.Sprintf("[%s][ %s ][-] %s\n", usr_color, username, text)
				}
				messageView.Write([]byte(formated_str))
				inputChan <- text;
				messageView.ScrollToEnd()
				inputField.SetText("")
			}
		}
	})
	inputField.SetBorder(true)
	app.SetFocus(inputField)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			app.SetFocus(inputField)
			return nil
		}

		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
			return nil
		}

		return event
	})

	// Vertical right layout (messages + input)
	rightPane := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(messageView, 0, 1, false).
		AddItem(inputField, 3, 0, true)

	// Horizontal layout: left (chatList), right (chat area)
	mainLayout := tview.NewFlex().
		AddItem(chatList, 30, 0, true).     // Leave space for future "active users"
		AddItem(rightPane, 0, 2, false)

	// Simulate incoming messages
	go func() {
		for msg := range outputChan {
			app.QueueUpdateDraw(func() {
				messageView.Write([]byte(msg + "\n"));
			});
		}
	}()

	// Run
	if err := app.SetRoot(mainLayout, true).EnableMouse(true).Run(); err != nil {
		panic(err);
	}
}

func handle_server_read(sockfd net.Conn, key []byte) {
	buffer := make([]byte, 4096);

	for {
		size, err := sockfd.Read(buffer);
		if err != nil {
			outputChan <- fmt.Sprintf("[!] Error: %v\n", err);
			close(outputChan);
			handle_error();
			break;
		}

		var pkt axion.Packet;
		if err := json.Unmarshal(buffer[:size], &pkt); err != nil {
			outputChan <- fmt.Sprintf("[!] Error: %v\n", err);
			close(outputChan);
			handle_error();
			break;
		}

		color := axion.Get_color(pkt.Sender);

		if !pkt.Encrypted {
			if pkt.Sender == "SERVER" {
				if data, found := strings.CutPrefix(pkt.Data, "CLIENTS "); found {
					nClients := strings.Fields(data);
					if len(nClients) < 1 {
						continue;
					}

					app.QueueUpdateDraw(func() {
						chatList.Clear()
						for _, client := range nClients {
							chatList.AddItem(client, "", 0, nil)
						}
					})
					continue;
				}
			}

			outputChan <- fmt.Sprintf("[red][ %s ][-] %s", pkt.Sender, pkt.Data);

			if strings.Contains(pkt.Data, "joined") || strings.Contains(pkt.Data, "left") {
				client_list(sockfd, username);
			}

			continue;
		}

		decrypted_data, err := pkt.Decrypt_data(key);
		if err != nil {
			outputChan <- fmt.Sprintf("[!] Error: %v\n", err);
			handle_error();
		}

		if pkt.Reciever == "SERVER" {
			outputChan <- fmt.Sprintf("[%s][ %s ][-] %s", color, pkt.Sender, decrypted_data);
		} else {
			outputChan <- fmt.Sprintf("[%s][ %s (private) ][-] %s", color, pkt.Sender, decrypted_data);
		}
	}
}

func handle_server_write(sockfd net.Conn, key []byte) {
	for msg := range inputChan {
		reciever := "SERVER";

		{
			/* Parse Input
			 * TODO: add /help
			*/

			if data, found := strings.CutPrefix(msg, "/msg "); found {
				x := strings.Split(data, " ");
				if len(x) < 2 {
					continue;
				}
				reciever = x[0]; // check len
				msg = strings.Join(x[1:], " ");
			}

			if _, found := strings.CutPrefix(msg, "/exit"); found {
				handle_error();
				break;
			}
		}

		pkt := axion.New(username, reciever);
		pkt.Set_data(key, msg);

		encoded, err := json.Marshal(pkt);
		if err != nil {
			fmt.Printf("\n[!] Error marshalling packet!\n");
			continue;
		}

		if _, err := sockfd.Write(encoded); err != nil {
			fmt.Printf("\n[!] Error writting to socket!\n");
			return;
		}
	}
}

func client_list(fd net.Conn, username string) {
	pkt := axion.New(username, "SERVER");
	pkt.Data = "CLIENTS";
	encoded, err := json.Marshal(pkt);
	if err != nil {
		fmt.Printf("[!] Error marshalling packet!\n");
		return;
	}

	if _, err := fd.Write(encoded); err != nil {
		fmt.Printf("[!] Error writting to socket!\n");
		return;
	}
}

func handle_error() {
	app.QueueUpdateDraw(func() {
		app.Stop();
	})
}
