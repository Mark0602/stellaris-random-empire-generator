package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type App struct {
	ctx context.Context
	rng *rand.Rand
}

func NewApp() *App {
	return &App{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

type GenerateRequest struct {
	Mode              string   `json:"mode"`
	PlayerCount       int      `json:"playerCount"`
	PlayerNames       []string `json:"playerNames"`
	EnabledDLCs       []string `json:"enabledDLCs"`
	TeamAOriginBans   []string `json:"teamAOriginBans"`
	TeamBOriginBans   []string `json:"teamBOriginBans"`
	TeamASpeciesBans  []string `json:"teamASpeciesBans"`
	TeamBSpeciesBans  []string `json:"teamBSpeciesBans"`
	AllowCrisisBuilds bool     `json:"allowCrisisBuilds"`
	AllowNomadicTrait bool     `json:"allowNomadicTrait"`
	RandomizeRulesSet bool     `json:"randomizeRulesSet"`
	RandomizeOrigins  bool     `json:"randomizeOrigins"`
	RandomizeSpecies  bool     `json:"randomizeSpecies"`
	RandomizeCivics   bool     `json:"randomizeCivics"`
	RandomizeTraits   bool     `json:"randomizeTraits"`
}

type GenerateResponse struct {
	Empires []Empire `json:"empires"`
	Notes   []string `json:"notes"`
}

type OptionList struct {
	DLCs    []DLCOption `json:"dlcs"`
	Origins []string    `json:"origins"`
	Species []string    `json:"species"`
}

type TeamRollRequest struct {
	PlayerNames []string `json:"playerNames"`
}

type TeamRoll struct {
	TeamA []string `json:"teamA"`
	TeamB []string `json:"teamB"`
}

type RerollAspectRequest struct {
	Rules  GenerateRequest `json:"rules"`
	Empire Empire          `json:"empire"`
	Aspect string          `json:"aspect"`
}

type Empire struct {
	Player       string   `json:"player"`
	Team         string   `json:"team"`
	Origin       string   `json:"origin"`
	Species      string   `json:"species"`
	Ethics       []string `json:"ethics"`
	Authority    string   `json:"authority"`
	Civics       []string `json:"civics"`
	Traits       []string `json:"traits"`
	Homeworld    string   `json:"homeworld"`
	PlaystyleTag string   `json:"playstyleTag"`
}

func (a *App) GetOptions() OptionList {
	return a.GetOptionsForDLCs(allDLCIDs())
}

func (a *App) GetOptionsForDLCs(enabledDLCs []string) OptionList {
	enabled := enabledDLCSet(enabledDLCs)
	return OptionList{
		DLCs:    dlcs,
		Origins: optionNames(filterByDLC(origins, enabled)),
		Species: optionNames(filterByDLC(speciesClasses, enabled)),
	}
}

func (a *App) GenerateEmpireSet(req GenerateRequest) GenerateResponse {
	normalizeRequest(&req)

	enabled := enabledDLCSet(req.EnabledDLCs)
	bannedOrigins := setFromSlices(req.TeamAOriginBans, req.TeamBOriginBans)
	bannedSpecies := setFromSlices(req.TeamASpeciesBans, req.TeamBSpeciesBans)

	originPool := filterOptions(filterByDLC(origins, enabled), bannedOrigins, func(option gameOption) bool {
		return !option.PlayerBlocked
	})
	speciesPool := filterOptions(filterByDLC(speciesClasses, enabled), bannedSpecies, func(option gameOption) bool {
		return !option.PlayerBlocked
	})
	if len(originPool) == 0 {
		originPool = filterByDLC(origins, enabled)
	}
	if len(speciesPool) == 0 {
		speciesPool = filterByDLC(speciesClasses, enabled)
	}

	notes := validatePool(req, originPool, speciesPool)
	if len(originPool) == 0 {
		originPool = filterByDLC(origins, enabled)
	}
	if len(speciesPool) == 0 {
		speciesPool = filterByDLC(speciesClasses, enabled)
	}
	empires := make([]Empire, 0, req.PlayerCount)
	players := a.playersForRequest(req)

	for i := 0; i < req.PlayerCount; i++ {
		origin := a.pick(originPool)
		species := a.pick(speciesPool)
		ethics := a.pickEthics(enabled)
		authority := a.pickAuthority(ethics, enabled)
		civics := a.pickCivics(authority, req.AllowCrisisBuilds, enabled)
		traits := a.pickTraits(species.Name, req.AllowNomadicTrait, enabled)

		empires = append(empires, Empire{
			Player:       players[i],
			Team:         teamName(req.Mode, i),
			Origin:       randomOrManual(req.RandomizeOrigins, origin.Name, "Manual Origin"),
			Species:      randomOrManual(req.RandomizeSpecies, species.Name, "Manual Species"),
			Ethics:       ethics,
			Authority:    authority,
			Civics:       randomListOrManual(req.RandomizeCivics, civics, []string{"Manual Civic 1", "Manual Civic 2"}),
			Traits:       randomListOrManual(req.RandomizeTraits, traits, []string{"Manual Traits"}),
			Homeworld:    a.pick(filterByDLC(homeworlds, enabled)).Name,
			PlaystyleTag: a.pickPlaystyleTag(origin),
		})
	}

	return GenerateResponse{
		Empires: empires,
		Notes:   notes,
	}
}

func (a *App) RerollEmpireAspect(req RerollAspectRequest) Empire {
	rules := req.Rules
	normalizeRequest(&rules)

	enabled := enabledDLCSet(rules.EnabledDLCs)
	bannedOrigins := setFromSlices(rules.TeamAOriginBans, rules.TeamBOriginBans)
	bannedSpecies := setFromSlices(rules.TeamASpeciesBans, rules.TeamBSpeciesBans)
	originPool := filterOptions(filterByDLC(origins, enabled), bannedOrigins, func(option gameOption) bool {
		return !option.PlayerBlocked
	})
	speciesPool := filterOptions(filterByDLC(speciesClasses, enabled), bannedSpecies, func(option gameOption) bool {
		return !option.PlayerBlocked
	})

	empire := req.Empire
	switch strings.ToLower(strings.TrimSpace(req.Aspect)) {
	case "origin":
		origin := a.pick(originPool)
		empire.Origin = randomOrManual(rules.RandomizeOrigins, origin.Name, "Manual Origin")
		empire.PlaystyleTag = a.pickPlaystyleTag(origin)
	case "species":
		species := a.pick(speciesPool)
		empire.Species = randomOrManual(rules.RandomizeSpecies, species.Name, "Manual Species")
	case "ethics":
		empire.Ethics = a.pickEthics(enabled)
	case "authority":
		empire.Authority = a.pickAuthority(empire.Ethics, enabled)
	case "civics":
		empire.Civics = randomListOrManual(rules.RandomizeCivics, a.pickCivics(empire.Authority, rules.AllowCrisisBuilds, enabled), []string{"Manual Civic 1", "Manual Civic 2"})
	case "traits":
		empire.Traits = randomListOrManual(rules.RandomizeTraits, a.pickTraits(empire.Species, rules.AllowNomadicTrait, enabled), []string{"Manual Traits"})
	case "homeworld":
		empire.Homeworld = a.pick(filterByDLC(homeworlds, enabled)).Name
	}
	return empire
}

func (a *App) RollTeams(req TeamRollRequest) TeamRoll {
	players := cleanedPlayerNames(req.PlayerNames)
	for len(players) < 4 {
		players = append(players, playerName(len(players)))
	}
	a.rng.Shuffle(len(players), func(i, j int) {
		players[i], players[j] = players[j], players[i]
	})
	return TeamRoll{
		TeamA: append([]string{}, players[:2]...),
		TeamB: append([]string{}, players[2:4]...),
	}
}

func normalizeRequest(req *GenerateRequest) {
	if req.Mode == "" {
		req.Mode = "playthrough"
	}
	if req.PlayerCount <= 0 {
		req.PlayerCount = 1
	}
	if req.Mode == "2v2" {
		req.PlayerCount = 4
	}
	if len(req.EnabledDLCs) == 0 {
		req.EnabledDLCs = allDLCIDs()
	}
	if !req.RandomizeRulesSet {
		req.RandomizeOrigins = true
		req.RandomizeSpecies = true
		req.RandomizeCivics = true
		req.RandomizeTraits = true
	}
}

func validatePool(req GenerateRequest, originPool []gameOption, speciesPool []gameOption) []string {
	notes := []string{}
	if len(req.TeamAOriginBans) > 2 || len(req.TeamBOriginBans) > 2 {
		notes = append(notes, "2v2 szabaly: csapatonkent legfeljebb 2 Origin ban legyen.")
	}
	if len(req.TeamASpeciesBans) > 1 || len(req.TeamBSpeciesBans) > 1 {
		notes = append(notes, "2v2 szabaly: csapatonkent legfeljebb 1 Species ban legyen.")
	}
	if len(originPool) == 0 {
		notes = append(notes, "Nincs elerheto Origin a banok utan, ezert a teljes listabol porgottem.")
	}
	if len(speciesPool) == 0 {
		notes = append(notes, "Nincs elerheto Species a banok utan, ezert a teljes listabol porgottem.")
	}
	if !req.AllowCrisisBuilds {
		notes = append(notes, "Crisis jellegu civic-ek ki vannak zarva.")
	}
	if !req.AllowNomadicTrait {
		notes = append(notes, "Nomadic trait ki van zarva.")
	}
	if !req.RandomizeOrigins {
		notes = append(notes, "Origin randomizalas kikapcsolva: manual pick.")
	}
	if !req.RandomizeSpecies {
		notes = append(notes, "Species randomizalas kikapcsolva: manual pick.")
	}
	if !req.RandomizeCivics {
		notes = append(notes, "Civic randomizalas kikapcsolva: manual pick.")
	}
	if !req.RandomizeTraits {
		notes = append(notes, "Trait randomizalas kikapcsolva: manual pick.")
	}
	return notes
}

func (a *App) pick(pool []gameOption) gameOption {
	if len(pool) == 0 {
		return gameOption{Name: "Nincs elerheto opcio", DLC: baseGameDLC}
	}
	return pool[a.rng.Intn(len(pool))]
}

func (a *App) pickPlaystyleTag(origin gameOption) string {
	if origin.DLC == "nomads" {
		return a.pick(nomadPlaystyleTags).Name
	}
	return a.pick(playstyleTags).Name
}

func (a *App) pickEthics(enabled map[string]bool) []string {
	if (enabled["utopia"] || enabled["synthetic-dawn"]) && a.rng.Intn(5) == 0 {
		return []string{"Gestalt Consciousness"}
	}

	axes := [][]string{
		{"Militarist", "Pacifist"},
		{"Xenophile", "Xenophobe"},
		{"Materialist", "Spiritualist"},
		{"Authoritarian", "Egalitarian"},
	}
	a.shuffleAxes(axes)

	if a.rng.Intn(2) == 0 {
		return []string{"Fanatic " + axes[0][a.rng.Intn(2)], axes[1][a.rng.Intn(2)]}
	}
	return []string{axes[0][a.rng.Intn(2)], axes[1][a.rng.Intn(2)], axes[2][a.rng.Intn(2)]}
}

func (a *App) pickAuthority(ethics []string, enabled map[string]bool) string {
	if contains(ethics, "Gestalt Consciousness") {
		gestaltAuthorities := []gameOption{{Name: "Hive Mind", DLC: "utopia"}, {Name: "Machine Intelligence", DLC: "synthetic-dawn"}}
		return a.pick(filterByDLC(gestaltAuthorities, enabled)).Name
	}
	return a.pick(filterByDLC(authorities, enabled)).Name
}

func (a *App) pickCivics(authority string, allowCrisis bool, enabled map[string]bool) []string {
	pool := filterOptions(filterByDLC(civics, enabled), nil, func(option gameOption) bool {
		if option.CrisisLike && !allowCrisis {
			return false
		}
		if authority == "Machine Intelligence" {
			return option.ForMachine
		}
		if authority == "Hive Mind" {
			return option.ForHive
		}
		return option.ForRegular
	})

	a.shuffleOptions(pool)
	if len(pool) < 2 {
		return []string{"Nincs eleg civic", "Kapcsolj be tobb DLC-t"}
	}
	return []string{pool[0].Name, pool[1].Name}
}

func (a *App) pickTraits(species string, allowNomadic bool, enabled map[string]bool) []string {
	pool := filterOptions(filterByDLC(traits, enabled), nil, func(option gameOption) bool {
		if option.Name == "Nomadic" && !allowNomadic {
			return false
		}
		if species == "Machine" || species == "Robot" {
			return option.ForMachine
		}
		return option.ForRegular
	})
	a.shuffleOptions(pool)
	if len(pool) < 3 {
		return []string{"Nincs eleg trait"}
	}
	return []string{pool[0].Name, pool[1].Name, pool[2].Name}
}

func (a *App) shuffleAxes(values [][]string) {
	a.rng.Shuffle(len(values), func(i, j int) {
		values[i], values[j] = values[j], values[i]
	})
}

func (a *App) shuffleOptions(values []gameOption) {
	a.rng.Shuffle(len(values), func(i, j int) {
		values[i], values[j] = values[j], values[i]
	})
}

func optionNames(values []gameOption) []string {
	names := make([]string, len(values))
	for i, value := range values {
		names[i] = value.Name
	}
	return names
}

func allDLCIDs() []string {
	ids := make([]string, len(dlcs))
	for i, dlc := range dlcs {
		ids[i] = dlc.ID
	}
	return ids
}

func enabledDLCSet(enabledDLCs []string) map[string]bool {
	enabled := map[string]bool{baseGameDLC: true}
	for _, dlc := range enabledDLCs {
		if dlc != "" {
			enabled[dlc] = true
		}
	}
	return enabled
}

func filterByDLC(values []gameOption, enabled map[string]bool) []gameOption {
	filtered := []gameOption{}
	for _, value := range values {
		dlc := value.DLC
		if dlc == "" {
			dlc = baseGameDLC
		}
		if enabled[dlc] {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func filterOptions(values []gameOption, banned map[string]bool, keep func(gameOption) bool) []gameOption {
	filtered := []gameOption{}
	for _, value := range values {
		if banned != nil && banned[value.Name] {
			continue
		}
		if keep != nil && !keep(value) {
			continue
		}
		filtered = append(filtered, value)
	}
	return filtered
}

func setFromSlices(slices ...[]string) map[string]bool {
	set := map[string]bool{}
	for _, values := range slices {
		for _, value := range values {
			if value != "" {
				set[value] = true
			}
		}
	}
	return set
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func (a *App) playersForRequest(req GenerateRequest) []string {
	players := cleanedPlayerNames(req.PlayerNames)
	for len(players) < req.PlayerCount {
		players = append(players, playerName(len(players)))
	}
	players = append([]string{}, players[:req.PlayerCount]...)
	if req.Mode == "2v2" {
		a.rng.Shuffle(len(players), func(i, j int) {
			players[i], players[j] = players[j], players[i]
		})
	}
	return players
}

func cleanedPlayerNames(values []string) []string {
	players := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		name := strings.TrimSpace(value)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		players = append(players, name)
	}
	return players
}

func playerName(index int) string {
	return fmt.Sprintf("Player %d", index+1)
}

func randomOrManual(randomize bool, value string, manualLabel string) string {
	if !randomize {
		return manualLabel
	}
	return value
}

func randomListOrManual(randomize bool, values []string, manualLabels []string) []string {
	if !randomize {
		return manualLabels
	}
	return values
}

func teamName(mode string, index int) string {
	if mode != "2v2" {
		return "Solo"
	}
	if index < 2 {
		return "Team A"
	}
	return "Team B"
}
