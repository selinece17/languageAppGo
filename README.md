# 🌍 AI Language Tutor

A desktop app that helps you learn a new language by giving you sentences to translate and using Google's Gemini AI to check your answers, give feedback, and track what you struggle with.

---

## What Does This App Do?

This is a **desktop GUI application** (meaning it opens as a real window on your computer, not in a browser) that teaches you a foreign language through translation practice.

Here's the basic flow:

1. You create a profile with your name and a little avatar symbol
2. You pick a language (Spanish, French, Japanese, etc.) and a difficulty level
3. The AI generates an English sentence for you to translate
4. You type your translation and hit "Check My Answer"
5. The AI grades your answer, tells you if you were right or wrong, and explains any mistakes
6. Wrong answers are automatically saved to your personal vocabulary list so you can review them later
7. Your progress (accuracy, streak days, weak spots) is tracked over time

**No subscription or paid account is required.** The app uses Google's Gemini AI which has a free tier that is more than enough for personal practice.

---

## What You Need Before You Start

Before you can run this app you need three things:

| Thing | What it is | Cost |
|---|---|---|
| **Go** | The programming language the app is written in | Free |
| **A terminal** | A text-based window where you type commands | Already on your computer |
| **A Gemini API key** | A password that lets the app talk to Google's AI | Free |

Don't worry if you've never used a terminal before — this guide walks you through every command.

### What is a terminal?

A terminal (also called a "command prompt" or "shell") is a window where you type instructions to your computer instead of clicking. It looks old-fashioned but it's how developers run programs.

- **Windows**: Press `Win + R`, type `cmd`, press Enter. Or search for "Command Prompt" in the Start menu. Even better: search for "PowerShell".
- **Mac**: Press `Cmd + Space`, type `Terminal`, press Enter.
- **Linux**: Press `Ctrl + Alt + T`, or search for "Terminal" in your apps.

---

## Step 1 — Install Go

Go is the programming language this app is written in. You need it installed so your computer can understand and run the code.

### Check if Go is already installed

Open your terminal and type this command, then press Enter:

```
go version
```

If you see something like `go version go1.22.0 ...` then Go is already installed and you can skip to Step 2.

If you see an error like `command not found` or `'go' is not recognized`, then you need to install it.

### Installing Go

1. Go to **https://go.dev/dl/**
2. Click the big blue download button for your operating system (Windows, macOS, or Linux)
3. Open the downloaded file and follow the installer — just click Next/Continue/Install on every screen
4. **Close your terminal and open a new one** (important — the terminal needs to restart to detect the new installation)
5. Type `go version` again to confirm it worked

> **What is Go?** Go (also called Golang) is a programming language created by Google. It's known for being fast and easy to read. This app is written in Go.

---

## Step 2 — Get a Free Gemini API Key

The app needs to talk to Google's Gemini AI to generate questions and grade your answers. To do that it needs an **API key** — think of it like a password that proves to Google's servers that requests are coming from you.

### How to get your free key

1. Go to **https://aistudio.google.com/**
2. Sign in with any Google account (Gmail, etc.)
3. Click **"Get API key"** in the left sidebar
4. Click **"Create API key"**
5. Select **"Create API key in new project"**
6. A long string starting with `AIza` will appear — that's your key!
7. Click the copy button next to it

> **Keep your key private.** Don't post it publicly on GitHub or share it with people you don't trust. Anyone with your key can use your free quota.

### Which model should I use?

This app is built and tested with **`gemini-2.5-flash`** — that's the recommended free model. When you create your API key in Google AI Studio, make sure the project has access to Gemini 2.5 Flash. You don't need to select a model anywhere in the app — it's already configured to use it.

### Is it really free?

Yes. Google's free tier for Gemini allows a generous number of requests per day — far more than you'd ever use just practising vocabulary. You do not need to enter a credit card.

---

## Step 3 — Download the Code

You need to get the app's source code onto your computer.

### If you have Git installed

Git is a tool for downloading and managing code. If you have it, run this in your terminal:

```
git clone https://github.com/your-username/languageapp.git
```

Then move into the folder it created:

```
cd languageapp
```

### If you don't have Git

1. Go to the project page on GitHub
2. Click the green **"Code"** button
3. Click **"Download ZIP"**
4. Unzip the downloaded file somewhere easy to find (like your Desktop or Documents folder)
5. In your terminal, navigate to that folder:

```
# On Mac or Linux:
cd ~/Desktop/languageapp

# On Windows (adjust the path to wherever you unzipped it):
cd C:\Users\YourName\Desktop\languageapp
```

> **What does `cd` mean?** It stands for "change directory." It's how you navigate between folders in the terminal. Think of it like double-clicking a folder, but with text.

---

## Step 4 — Install Dependencies

This app uses a few external libraries (pre-written chunks of code made by other developers). The main one is **Fyne**, which is what draws the windows and buttons on screen.

You only need to run **one command** to download everything automatically. Make sure you're inside the project folder (from Step 3), then run:

```
go mod tidy
```

This will:
- Read the list of required libraries from a file called `go.mod`
- Download all of them from the internet
- Store them on your computer so they're ready to use

It might take a minute or two the first time. You'll see some output lines as it downloads things. That's normal.

> **What is a dependency?** A dependency is code written by someone else that your project relies on. Instead of writing every single feature from scratch, programmers reuse libraries. `go mod tidy` is like telling Go "go fetch everything this project needs."

### Extra requirement on Linux

If you're on Linux, Fyne needs a few system-level graphics libraries. Run this first:

```
# Ubuntu / Debian / Linux Mint:
sudo apt-get install gcc libgl1-mesa-dev xorg-dev

# Fedora:
sudo dnf install gcc mesa-libGL-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel
```

> You don't need to do anything extra on Windows or Mac.

---

## Step 5 — Run the App

You're ready to go! In your terminal (still inside the project folder), run:

```
go run .
```

A window titled **"🌍 AI Language Tutor"** should appear on your screen.

> **What does `go run .` do?** It tells Go to compile (translate your source code into something the computer can execute) and immediately run the program. The `.` means "use the code in the current folder."


## How to Use the App

### First launch — Create a Profile

When the app opens you'll see the **Profile Screen**. Since it's your first time, there are no profiles yet.

1. In the **"Create New Profile"** section, type your name in the Name field
2. Click one of the avatar buttons (`[*]`, `[@]`, `[#]`, etc.) to pick your symbol
3. Click **"Create Profile"**

You'll be taken to the Home Screen automatically.

> You can create multiple profiles — handy if several people share the same computer. Each profile has its own completely separate progress and vocabulary list.

---

### Home Screen — Set Up a Session

This is your control centre. Here you'll:

#### 1. Enter your API Key
Paste the Gemini API key you copied in Step 2 into the **"Google Gemini API Key"** field and click **"Save Key"**.

You only need to do this once — the key is saved to your computer and will be there next time you open the app.

> **Model note:** This app uses **`gemini-2.5-flash`**, which is free. Make sure your API key was created in [Google AI Studio](https://aistudio.google.com/) (not Google Cloud Console) to ensure free tier access.

#### 2. Choose a Language
Click the **Language** dropdown and pick what you want to practise. Options include:
Spanish, French, German, Italian, Portuguese, Japanese, Chinese, Korean, Arabic, Russian.

#### 3. Choose a Difficulty
- **Beginner** — Short, simple sentences with common vocabulary. Good if you're just starting out.
- **Intermediate** — More complex grammar and less common words.
- **Advanced** — Challenging sentences that test nuanced vocabulary and grammar.

#### 4. Start Practising
Click the big **"Start Practicing"** button.

---

### Practice Screen — Translate Sentences

This is where the learning happens.

1. The AI generates an English sentence and shows it in the **"Translate this sentence"** box
2. Type your translation in the text area below it
3. Click **"Check My Answer"**
4. Wait a moment while the AI grades your response
5. The **AI Feedback** card will appear, showing:
   - `[CORRECT]` or `[WRONG]` so you know immediately
   - A brief explanation of what was right or what you got wrong
   - The correct translation (so you can learn from mistakes)
6. Click **"Next Question"** to get a new sentence

Your **score** in the top right updates after every answer, showing your accuracy for the current session.

> **Wrong answers are saved automatically.** Any sentence you get wrong is added to your Vocabulary List so you can review it later.

---

### Vocabulary List Screen

Click **"My Vocab List"** from the Practice Screen or **"My Vocabulary List"** from the Home Screen.

This shows every sentence you've gotten wrong, along with:
- What you typed (your wrong answer)
- What the correct answer was
- Which language it's for
- When it was added and how many times you've missed it

You can:
- Click **"Remove"** on any entry to delete just that one
- Click **"Clear Vocab List"** to wipe everything and start fresh

---

### Progress Screen

Click **"View My Progress"** from the Home Screen.

Shows three sections:

**Overall Statistics**
Your total attempts, how many were correct, your overall accuracy percentage, and how many different days you've practised.

**Weak Spots**
Areas where you've made the most mistakes. Over time this tells you what to focus on.

**Recent Attempts (last 10)**
A quick log of your most recent questions so you can see how the current session went.

You can also click **"Clear All Progress"** to reset everything back to zero (this cannot be undone).

---

### Switching Profiles

From the Home Screen, click **"Switch Profile"** to go back to the profile picker. You can select a different user, create a new one, or delete an old one (deleting also removes all that person's data).

---

## Project File Structure

Here's what each file in the project does, in plain English:

```
languageapp/
│
├── main.go               Starting point — opens the window and launches the app
│
├── go.mod                Lists the external libraries the project needs
├── go.sum                Security checksums for those libraries (don't edit this)
│
└── ui/                   All the visual screens and interface code
│   ├── main_ui.go        The "hub" — manages navigation between screens
│   ├── home_screen.go    The setup screen (API key, language, difficulty)
│   ├── practice_screen.go  The translation exercise screen
│   ├── profile_screen.go   The profile selection / creation screen
│   ├── progress_screen.go  The stats and history screen
│   ├── vocab_screen.go     The vocabulary review screen
│   └── theme.go          Defines the teal colour scheme
│
└── models/
│   └── models.go         Defines the data shapes (what a Profile looks like, etc.)
│
└── storage/
│   └── storage.go        Handles reading and writing data to disk
│
└── claude/
    └── client.go         Handles communication with the Gemini AI API
```

> **Why are there so many files?** In programming, it's good practice to split code into separate files based on what each part does. This is called "separation of concerns" — it makes the code easier to read, fix, and build on top of.

---

## Where Your Data Is Saved

The app saves all your data automatically. You don't need to do anything. Here's where it goes:

| Operating System | Location |
|---|---|
| **Windows** | `C:\Users\YourName\AppData\Roaming\languageapp\` |
| **macOS** | `~/Library/Application Support/languageapp/` |
| **Linux** | `~/.config/languageapp/` |

Inside that folder:

```
languageapp/
├── settings.json       Your API key and last-used language/difficulty
├── profiles.json       List of all profiles
└── users/
    └── yourname/
        ├── progress.json   All your attempt history and weak spots
        └── vocab.json      Your personal vocabulary list
```

These are plain JSON files — you can open them in any text editor if you're curious what they look like. JSON is just a structured text format used to store data.

> **Backing up your data:** If you want to save your progress before reinstalling the app or switching computers, just copy that whole `languageapp` folder to a safe place.

---

## Supported Languages

| Language | Code |
|---|---|
| Spanish | es |
| French | fr |
| German | de |
| Italian | it |
| Portuguese | pt |
| Japanese | ja |
| Chinese | zh |
| Korean | ko |
| Arabic | ar |
| Russian | ru |

---

## Troubleshooting Common Problems

### "command not found: go" or "'go' is not recognized"

Go isn't installed, or the terminal doesn't know where to find it.

- Make sure you completed the Go installation from Step 1
- **Close your terminal and open a brand new one** — this is the most common fix
- If it still doesn't work, try restarting your computer

---

### The app window doesn't open (no error message)

- Make sure you're running the command from inside the project folder
- Try running `go build .` first. If there's an error it will be printed to the terminal.

---

### "Invalid Key Format" when saving API key

Your key must start with `AIza`. Make sure you copied the full key from Google AI Studio — sometimes copy-paste can accidentally cut off the beginning or end.

---

### "API error 400" or "API error 403"

- **400**: The key format is wrong. Double-check you copied the whole thing.
- **403**: The key exists but doesn't have permission. Make sure you created the key in Google AI Studio (aistudio.google.com), not Google Cloud Console. They look similar but are different.

---

### "request failed — check your internet connection"

Exactly what it says — the app couldn't reach Google's servers. Check that:
- You're connected to the internet
- Nothing is blocking the app (firewall, VPN, corporate network)

---

### Questions load very slowly

The Gemini AI model usually responds in 1–5 seconds. If it's consistently slower, it could be:
- A slow internet connection
- Google's servers being temporarily busy — just wait and try again

---

### On Linux: window doesn't open / graphics error

You probably need the system libraries from Step 4. Run the `apt-get` or `dnf` install command from that section, then try again.

---

### "a profile named 'X' already exists"

Profile names must be unique. Either pick a different name or delete the existing profile first from the profile screen.
------------------------------------------------------------------------

# License

This project is intended for educational purposes.
