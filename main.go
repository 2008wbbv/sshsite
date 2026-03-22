package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
)

const (
	host    = "0.0.0.0"
	port    = 2222
	keyPath = ".ssh/server_key"
)

// ── colors (Catppuccin Mocha) ─────────────────────────────────────────────────

var (
	green  = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1"))
	blue   = lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa"))
	mauve  = lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7"))
	peach  = lipgloss.NewStyle().Foreground(lipgloss.Color("#fab387"))
	yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af"))
	teal   = lipgloss.NewStyle().Foreground(lipgloss.Color("#94e2d5"))
	red    = lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8"))
	dim    = lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70"))
	dim2   = lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086"))
	bold   = lipgloss.NewStyle().Bold(true)
	fg     = lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4"))
)

func g(s string) string  { return green.Render(s) }
func b(s string) string  { return blue.Render(s) }
func m(s string) string  { return mauve.Render(s) }
func p(s string) string  { return peach.Render(s) }
func y(s string) string  { return yellow.Render(s) }
func t(s string) string  { return teal.Render(s) }
func r(s string) string  { return red.Render(s) }
func d(s string) string  { return dim.Render(s) }
func d2(s string) string { return dim2.Render(s) }
func fw(s string) string { return bold.Render(s) }
func rule(n int) string  { return d(strings.Repeat("─", n)) }
func sec(title string) string {
	n := max(4, 46-len(title)-4)
	return "\n" + d("── "+title+" ") + rule(n)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ── content ───────────────────────────────────────────────────────────────────

func banner() string {
	lines := []string{
		blue.Render(" ██████╗ ███████╗███╗   ██╗"),
		blue.Render(" ██╔══██╗██╔════╝████╗  ██║"),
		blue.Render(" ██████╔╝█████╗  ██╔██╗ ██║"),
		blue.Render(" ██╔══██╗██╔══╝  ██║╚██╗██║"),
		blue.Render(" ██████╔╝███████╗██║ ╚████║"),
		blue.Render(" ╚═════╝ ╚══════╝╚═╝  ╚═══╝"),
	}
	out := strings.Join(lines, "\n")
	out += "\n"
	out += "\n" + fw(g("ben vaccaro")) + d2("  ·  ") + "builder · engineer · senior @ millis high"
	out += "\n" + d("millis, ma  ·  class of 2027  ·  he/him")
	out += "\n"
	out += "\n" + rule(46)
	out += "\n" + d("type ") + g("help") + d(" to get started  ·  ") + g("whoami") + d(" for the short version")
	return out
}

func helpCmd() string {
	out := sec("commands")
	out += "\n"
	cmds := [][2]string{
		{g("about"), "who i am"},
		{g("projects"), "things i've built"},
		{g("experience"), "where i've worked & led"},
		{g("skills"), "tools & technologies"},
		{g("awards"), "wins & recognition"},
		{g("contact"), "how to reach me"},
		{g("now"), "what i'm up to right now"},
		{g("whoami"), "the one-liner"},
		{g("ls"), "list sections"},
		{g("cat") + " " + d2("<section>"), "read a section directly"},
		{g("clear"), "clear the screen"},
		{g("exit"), "close the connection"},
	}
	for _, c := range cmds {
		out += "\n  " + c[0] + "  " + d(c[1])
	}
	out += "\n"
	out += "\n" + d("tip: ") + d2("↑↓") + d(" history  ·  ") + d2("tab") + d(" autocomplete")
	return out
}

func whoamiCmd() string {
	out := "\n" + fw("ben vaccaro") + d2(" — ") + "junior @ millis high. building hardware, robots, and software."
	out += "\n" + d2("  mit bwsi cubesat lead  ·  founder @ mpi  ·  sea perch top-50  ·  mit blueprint winner  ·  comp bio researcher")
	return out
}

func aboutCmd() string {
	out := sec("about")
	out += "\n"
	out += "\nI'm a junior at " + b("Millis High School") + " (class of 2027)"
	out += "\nfocused on hands-on engineering — the kind where you"
	out += "\nactually build the thing, break it, and fix it."
	out += "\n"
	out += "\nI lead the " + g("MIT BWSI CubeSat") + " team under MIT Lincoln Lab mentorship,"
	out += "\nrun " + p("Moore Performance Industries") + " (custom PCBs & electronics)"
	out += "\nout of my basement, and compete with " + t("Sea Perch / AUAVS") + " robotics."
	out += "\n"
	out += "\nOutside engineering: " + y("Boy Scouts") + " (SPL, prospective Eagle Scout),"
	out += "\n" + m("U.S. Naval Sea Cadet Corps") + " corpsman, Next Voters fellow,"
	out += "\nindependent comp bio researcher, DECA, honor roll."
	out += "\n"
	out += "\n" + d("gpa: ") + g("4.6 weighted") + d("  ·  all honors  ·  spanish immersion")
	out += "\n"
	out += "\n" + rule(46)
	return out
}

func projectsCmd() string {
	type project struct {
		name string
		year string
		col  func(string) string
		desc []string
		tech []string
		url  string
	}
	projects := []project{
		{
			name: "MIT Blueprint Hackathon — 1st Place Overall", year: "2026", col: g,
			desc: []string{
				"Arduino wearable tracking runner speed vs live MBTA train data.",
				"Won across all tracks. Built in 10 hours with a team of 4.",
			},
			tech: []string{"arduino", "python", "mbta api", "hardware"},
			url:  "github.com/2008wbbv/epic-blueprint-repo",
		},
		{
			name: "MBTA-Plus", year: "2025", col: b,
			desc: []string{"Live Boston subway data on an interactive map.", "Lightweight client-side web app."},
			tech: []string{"html", "javascript", "api"},
			url:  "github.com/2008wbbv/MBTA-Plus",
		},
		{
			name: "OrbitCLI", year: "2025", col: m,
			desc: []string{"CLI tool that renders orbiting planets as an ASCII focus timer."},
			tech: []string{"python", "cli"},
			url:  "github.com/2008wbbv/OrbitCLI",
		},
		{
			name: "CollegeExplorer", year: "2025", col: t,
			desc: []string{"Client-side app for real-time college data via the Scorecard API."},
			tech: []string{"html", "javascript", "rest api"},
			url:  "github.com/2008wbbv/CollegeExplorer",
		},
		{
			name: "CubeSat Shake Code", year: "2025", col: p,
			desc: []string{
				"Vibration + sensor code for MIT BWSI satellite.",
				"Part of autonomous navigation & fault tolerance research.",
			},
			tech: []string{"python", "embedded", "opencv"},
			url:  "github.com/2008wbbv/CubesatShakeCode",
		},
		{
			name: "Dream Venture Hackathon — 1st Place", year: "2025", col: y,
			desc: []string{"Prototype designed and validated under tight deadline."},
			tech: []string{"hardware", "rapid proto"},
		},
	}

	out := sec("projects")
	for _, proj := range projects {
		out += "\n"
		out += "\n  " + fw(proj.col(proj.name)) + d2("  "+proj.year)
		for _, desc := range proj.desc {
			out += "\n  " + d2(desc)
		}
		techStr := ""
		for _, tech := range proj.tech {
			techStr += "[" + tech + "] "
		}
		out += "\n  " + d2(techStr)
		if proj.url != "" {
			out += d("  ↗ ") + d2(proj.url)
		}
	}
	out += "\n"
	out += "\n  " + d("more: ") + b("github.com/2008wbbv")
	out += "\n"
	out += "\n" + rule(46)
	return out
}

func experienceCmd() string {
	type role struct {
		title string
		org   string
		dates string
		col   func(string) string
		items []string
	}
	roles := []role{
		{
			title: "CubeSat Team Lead", org: "MIT Lincoln Laboratory · BWSI", dates: "2025–present", col: b,
			items: []string{
				"Leading system design for a student-built satellite — autonomous nav & fault tolerance.",
				"Built computer vision pipeline (OpenCV) for obstacle detection in simulation.",
				"Coordinates subsystem integration across multidisciplinary team.",
			},
		},
		{
			title: "Founder & Head of Operations", org: "Moore Performance Industries", dates: "2024–present", col: p,
			items: []string{
				"Design & build custom electronics: schematic → PCB → shipped product.",
				"Custom enclosures, DIY kits, assembled builds. Based in greater Boston.",
				"moorepreformance.dev",
			},
		},
		{
			title: "Robotics Team Lead", org: "Sea Perch / AUAVS", dates: "2023–present", col: g,
			items: []string{
				"Redesigned underwater + aerial robot platforms for modularity.",
				"3D-printed mounting system improved vehicle stability.",
				"Team placed top 50 nationally in first year competing.",
			},
		},
		{
			title: "Co-Lead, Education & Curriculum", org: "Next Voters Fellowship", dates: "2025–present", col: m,
			items: []string{
				"Designing first-of-a-kind AI-first civic education curriculum.",
				"National cohort — youth voter engagement & policy initiatives.",
			},
		},
		{
			title: "Grant Lead", org: "Youth in Philanthropy", dates: "2024", col: y,
			items: []string{"Evaluated applications & allocated $10,000 in grant funding."},
		},
		{
			title: "Organizer", org: "Dream Venture Hackathon", dates: "2025–present", col: g,
			items: []string{
				"Organizing and running the Dream Venture Hackathon for student entrepreneurs.",
				"Also competed and won 1st place as a participant.",
			},
		},
		{
			title: "Assistant CPO / Corpsman", org: "U.S. Naval Sea Cadet Corps", dates: "2023–present", col: t,
			items: []string{
				"Oversaw training & readiness for 170 members.",
				"Primary corpsman — first aid, safety, emergency readiness.",
			},
		},
		{
			title: "Senior Patrol Leader / Prospective Eagle Scout", org: "Boy Scouts of America · NYLT", dates: "2023–present", col: y,
			items: []string{
				"Senior Patrol Leader — highest youth leadership position in the troop.",
				"Mentored patrol of youth leaders through week-long NYLT leadership course.",
			},
		},
		{
			title: "Independent Researcher", org: "Computational Biology & CS", dates: "2025–present", col: m,
			items: []string{
				"Independent research on AlphaFold protein structure prediction.",
				"Research on transistor miniaturization and alternative computing paradigms.",
				"Comp bio methods applied to CS research questions.",
			},
		},
	}

	out := sec("experience")
	for _, role := range roles {
		out += "\n"
		out += "\n  " + fw(role.col(role.title))
		out += "\n  " + d2(role.org+"  ·  "+role.dates)
		for _, item := range role.items {
			out += "\n  " + d("· ") + d2(item)
		}
	}
	out += "\n"
	out += "\n" + rule(46)
	return out
}

func skillsCmd() string {
	out := sec("skills")
	out += "\n"
	cats := [][2]string{
		{"languages", g("python") + d2(" · ") + b("html/css/js") + d2(" · ") + m("bash")},
		{"hardware", p("pcb design") + d2(" · ") + p("kicad/eagle") + d2(" · ") + p("3d printing") + d2(" · ") + p("soldering")},
		{"embedded", t("arduino") + d2(" · ") + t("esp32") + d2(" · ") + t("opencv") + d2(" · ") + t("sensors")},
		{"systems", y("linux") + d2(" · ") + y("git") + d2(" · ") + y("systems arch")},
		{"research", m("alphafold") + d2(" · ") + m("comp bio") + d2(" · ") + m("protein structure")},
	}
	for _, cat := range cats {
		label := cat[0]
		for len(label) < 12 {
			label += " "
		}
		out += "\n  " + d2(label) + cat[1]
	}
	out += "\n"
	out += "\n" + rule(46)
	return out
}

func awardsCmd() string {
	out := sec("awards")
	out += "\n"
	awards := [][3]string{
		{g("1st place"), "MIT Blueprint Hackathon (all tracks)", "2026"},
		{g("1st place"), "Dream Venture Hackathon", "2025"},
		{b("semifinalist"), "International Research Olympiad", "2024"},
		{b("honor roll"), "International Research Olympiad", "2024"},
		{m("cs student of the year"), "Millis High School", "2024"},
		{y("high honor roll"), "all As, all quarters", "2024"},
		{p("top 50 national"), "Sea Perch competition, first year", "2024"},
		{t("$10k allocated"), "Youth in Philanthropy grant lead", "2024"},
	}
	for _, a := range awards {
		out += "\n  " + a[0] + d2("  "+a[1]) + d("  "+a[2])
	}
	out += "\n"
	out += "\n" + rule(46)
	return out
}

func contactCmd() string {
	out := sec("contact")
	out += "\n"
	out += "\n  " + d2("email     ") + b("benvaccaro@proton.me")
	out += "\n  " + d2("github    ") + b("github.com/2008wbbv")
	out += "\n  " + d2("linkedin  ") + b("linkedin.com/in/benvac")
	out += "\n  " + d2("web       ") + b("moorepreformance.dev")
	out += "\n  " + d2("instagram ") + b("@bnvac")
	out += "\n"
	out += "\n  " + d("email is the best way. i respond within a day.")
	out += "\n"
	out += "\n" + rule(46)
	return out
}

func nowCmd() string {
	ts := time.Now().UTC().Format("2006-01-02 15:04 UTC")
	out := sec("now")
	out += "\n"
	out += "\n  " + d2("building   ") + fw(g("cubesat autonomous nav system"))
	out += "\n  " + d2("           ") + "MIT BWSI — fault tolerance + obstacle detection"
	out += "\n"
	out += "\n  " + d2("also built ") + fw(p("epic-blueprint-repo")) + d2("  ← most recent commit")
	out += "\n  " + d2("           ") + fw(t("lunar_mission_planner")) + d2("  ← this week")
	out += "\n"
	out += "\n  " + d2("running    ") + "MPI — new PCB designs, sourcing, shipping kits"
	out += "\n  " + d2("researching") + "comp bio — AlphaFold, transistor miniaturization"
	out += "\n  " + d2("studying   ") + "junior year · millis high"
	out += "\n"
	out += "\n  " + d(ts)
	out += "\n"
	out += "\n" + rule(46)
	return out
}

func lsCmd() string {
	sections := []string{"about", "projects", "experience", "skills", "awards", "contact", "now"}
	parts := make([]string, len(sections))
	for i, s := range sections {
		parts[i] = g(s)
	}
	return "\n" + strings.Join(parts, "  ")
}

// ── bubbletea model ───────────────────────────────────────────────────────────

type model struct {
	output  []string
	input   string
	history []string
	histIdx int
	width   int
	height  int
	ready   bool
}

func initialModel() model {
	m := model{histIdx: -1}
	m.output = append(m.output, banner())
	return m
}

type errMsg error

func (m model) Init() tea.Cmd {
	return nil
}

var sections = map[string]bool{
	"about": true, "projects": true, "experience": true,
	"skills": true, "awards": true, "contact": true, "now": true,
}

func (m model) runCommand(raw string) (model, tea.Cmd) {
	parts := strings.Fields(strings.TrimSpace(raw))
	if len(parts) == 0 {
		return m, nil
	}
	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	// echo the command
	prompt := "\n" + g("ben") + d2("@") + b("benvaccaro") + d(":") + mauve.Render("~") + d2("$ ") + fg.Render(raw)
	m.output = append(m.output, prompt)

	var result string
	switch cmd {
	case "help":
		result = helpCmd()
	case "about":
		result = aboutCmd()
	case "projects":
		result = projectsCmd()
	case "experience":
		result = experienceCmd()
	case "skills":
		result = skillsCmd()
	case "awards":
		result = awardsCmd()
	case "contact":
		result = contactCmd()
	case "now":
		result = nowCmd()
	case "whoami":
		result = whoamiCmd()
	case "ls":
		result = lsCmd()
	case "cat":
		if len(args) == 0 {
			result = "\n" + r("cat: missing operand") + d("  (try ") + g("ls") + d(")")
		} else if sections[args[0]] {
			switch args[0] {
			case "about":
				result = aboutCmd()
			case "projects":
				result = projectsCmd()
			case "experience":
				result = experienceCmd()
			case "skills":
				result = skillsCmd()
			case "awards":
				result = awardsCmd()
			case "contact":
				result = contactCmd()
			case "now":
				result = nowCmd()
			}
		} else {
			result = "\n" + r("cat: "+args[0]+": no such file") + d("  (try ") + g("ls") + d(")")
		}
	case "clear":
		m.output = nil
		return m, nil
	case "banner":
		result = banner()
	case "ssh":
		result = "\n" + d("you're already in via ssh. nice.")
	case "exit", "quit", "logout":
		return m, tea.Quit
	default:
		result = "\n" + r(cmd+": command not found") + d("  (try ") + g("help") + d(")")
	}

	m.output = append(m.output, result)
	return m, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			input := m.input
			if strings.TrimSpace(input) != "" {
				m.history = append([]string{input}, m.history...)
				m.histIdx = -1
			}
			m.input = ""
			var cmd tea.Cmd
			m, cmd = m.runCommand(input)
			return m, cmd

		case tea.KeyBackspace, tea.KeyDelete:
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}

		case tea.KeyUp:
			if len(m.history) > 0 {
				if m.histIdx < len(m.history)-1 {
					m.histIdx++
				}
				m.input = m.history[m.histIdx]
			}

		case tea.KeyDown:
			if m.histIdx > 0 {
				m.histIdx--
				m.input = m.history[m.histIdx]
			} else {
				m.histIdx = -1
				m.input = ""
			}

		case tea.KeyTab:
			partial := strings.ToLower(strings.TrimSpace(m.input))
			allCmds := []string{"help", "about", "projects", "experience", "skills", "awards", "contact", "now", "whoami", "ls", "cat", "clear", "exit", "banner", "ssh"}
			var matches []string
			for _, c := range allCmds {
				if strings.HasPrefix(c, partial) {
					matches = append(matches, c)
				}
			}
			if len(matches) == 1 {
				m.input = matches[0]
			} else if len(matches) > 1 {
				prompt := "\n" + g("ben") + d2("@") + b("benvaccaro") + d(":") + mauve.Render("~") + d2("$ ") + fg.Render(partial)
				m.output = append(m.output, prompt)
				parts := make([]string, len(matches))
				for i, match := range matches {
					parts[i] = g(match)
				}
				m.output = append(m.output, "\n"+strings.Join(parts, "  "))
			}

		case tea.KeyCtrlC, tea.KeyCtrlD:
			m.input = ""

		case tea.KeyCtrlL:
			m.output = nil

		case tea.KeyRunes:
			m.input += string(msg.Runes)
		}
	}
	return m, nil
}

func (m model) View() string {
	var sb strings.Builder

	for _, line := range m.output {
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	// prompt line
	sb.WriteString("\n")
	sb.WriteString(g("ben") + d2("@") + b("benvaccaro") + d(":") + mauve.Render("~") + d2("$ ") + fg.Render(m.input))

	return sb.String()
}

// ── server ────────────────────────────────────────────────────────────────────

func main() {
	// ensure .ssh dir exists for host key
	if err := os.MkdirAll(".ssh", 0700); err != nil {
		log.Fatal(err)
	}

	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithHostKeyPath(keyPath),
		wish.WithMiddleware(
			bm.Middleware(func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
				pty, _, active := s.Pty()
				if !active {
					fmt.Fprintln(s, "No active terminal, aborting")
					return nil, nil
				}
				m := initialModel()
				m.width = pty.Window.Width
				m.height = pty.Window.Height
				return m, []tea.ProgramOption{tea.WithAltScreen()}
			}),
			lm.Middleware(),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("SSH server listening on %s:%d", host, port)
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-done
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Fatal(err)
	}
}
