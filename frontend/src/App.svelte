<script lang="ts">
  import { onMount } from 'svelte';
  import {
    GenerateEmpireSet as wailsGenerateEmpireSet,
    GetOptionsForDLCs as wailsGetOptionsForDLCs,
    RerollEmpireAspect as wailsRerollEmpireAspect,
    RollTeams as wailsRollTeams,
  } from '../wailsjs/go/main/App.js';
  import { main } from '../wailsjs/go/models';

  type DlcOption = {
    id: string;
    name: string;
    type: string;
  };

  type OptionList = {
    dlcs: DlcOption[];
    origins: string[];
    species: string[];
  };

  type Empire = {
    player: string;
    team: string;
    origin: string;
    species: string;
    ethics: string[];
    authority: string;
    civics: string[];
    traits: string[];
    homeworld: string;
    playstyleTag: string;
  };

  type GenerateResponse = {
    empires: Empire[];
    notes: string[];
  };

  type GenerateRequest = {
    mode: string;
    playerCount: number;
    playerNames: string[];
    enabledDLCs: string[];
    teamAOriginBans: string[];
    teamBOriginBans: string[];
    teamASpeciesBans: string[];
    teamBSpeciesBans: string[];
    allowCrisisBuilds: boolean;
    allowNomadicTrait: boolean;
    randomizeRulesSet: boolean;
    randomizeOrigins: boolean;
    randomizeSpecies: boolean;
    randomizeCivics: boolean;
    randomizeTraits: boolean;
  };

  type TeamRoll = {
    teamA: string[];
    teamB: string[];
  };

  type EmpireField = {
    label: string;
    value: string;
  };

  const previewDlcs: DlcOption[] = [
    { id: 'base', name: 'Base Game', type: 'core' },
    { id: 'utopia', name: 'Utopia', type: 'expansion' },
    { id: 'federations', name: 'Federations', type: 'expansion' },
    { id: 'aquatics', name: 'Aquatics', type: 'species pack' },
    { id: 'machine-age', name: 'The Machine Age', type: 'expansion' },
    { id: 'season-10', name: 'Season 10 Scenarios', type: 'free update' },
    { id: 'nomads', name: 'Nomads', type: 'expansion' },
    { id: 'willpower', name: 'Willpower', type: 'expansion' },
  ];

  const previewOrigins = [
    { name: 'Prosperous Unification', dlc: 'base' },
    { name: 'Mechanist', dlc: 'utopia' },
    { name: 'Void Dwellers', dlc: 'federations' },
    { name: 'Ocean Paradise', dlc: 'aquatics' },
    { name: 'Cybernetic Creed', dlc: 'machine-age' },
    { name: 'The Sacred Path', dlc: 'nomads' },
    { name: 'Forever Cruise', dlc: 'nomads' },
    { name: 'Arkship Exodus', dlc: 'nomads' },
  ];

  const previewSpecies = [
    { name: 'Humanoid', dlc: 'base' },
    { name: 'Mammalian', dlc: 'base' },
    { name: 'Aquatic', dlc: 'aquatics' },
    { name: 'Cybernetic', dlc: 'machine-age' },
  ];

  let mode: 'playthrough' | '2v2' = '2v2';
  let playerCount = 1;
  let playerNamesText = 'Mark\nPeti\nDani\nBence';
  let allowCrisisBuilds = false;
  let allowNomadicTrait = false;
  let randomizeOrigins = true;
  let randomizeSpecies = true;
  let randomizeCivics = true;
  let randomizeTraits = true;
  let dlcs: DlcOption[] = [];
  let enabledDLCs: string[] = [];
  let origins: string[] = [];
  let species: string[] = [];
  let teamAOriginBans: string[] = [];
  let teamBOriginBans: string[] = [];
  let teamASpeciesBans: string[] = [];
  let teamBSpeciesBans: string[] = [];
  let result: GenerateResponse = { empires: [], notes: [] };
  let teamRoll: TeamRoll = { teamA: [], teamB: [] };
  let isLoading = false;
  let isRevealing = false;
  let revealRunId = 0;
  let revealTimers: number[] = [];
  let jokerUsedEmpireIndexes: number[] = [];
  let jokerPendingKey = '';
  let jokerAnimatingKey = '';
  const fieldCount = 7;

  onMount(async () => {
    const options = await getOptions([]);
    dlcs = options.dlcs;
    enabledDLCs = options.dlcs.map(dlc => dlc.id);
    await refreshOptions();
    await rollTeams();
  });

  async function refreshOptions(): Promise<void> {
    const options = await getOptions(enabledDLCs);
    origins = options.origins;
    species = options.species;
    teamAOriginBans = teamAOriginBans.filter(origin => origins.includes(origin));
    teamBOriginBans = teamBOriginBans.filter(origin => origins.includes(origin));
    teamASpeciesBans = teamASpeciesBans.filter(item => species.includes(item));
    teamBSpeciesBans = teamBSpeciesBans.filter(item => species.includes(item));
  }

  async function generate(): Promise<void> {
    const currentRunId = revealRunId + 1;
    revealRunId = currentRunId;
    clearRevealTimers();
    isLoading = true;
    isRevealing = false;
    jokerUsedEmpireIndexes = [];
    jokerPendingKey = '';
    jokerAnimatingKey = '';

    try {
      const nextResult = await generateEmpireSet({
        mode,
        playerCount,
        playerNames: playerNames(),
        enabledDLCs,
        teamAOriginBans,
        teamBOriginBans,
        teamASpeciesBans,
        teamBSpeciesBans,
        allowCrisisBuilds,
        allowNomadicTrait,
        randomizeRulesSet: true,
        randomizeOrigins,
        randomizeSpecies,
        randomizeCivics,
        randomizeTraits,
      });
      result = nextResult;
      teamRoll = {
        teamA: nextResult.empires.filter(empire => empire.team === 'Team A').map(empire => empire.player),
        teamB: nextResult.empires.filter(empire => empire.team === 'Team B').map(empire => empire.player),
      };
      isLoading = false;
      isRevealing = true;
      scheduleRevealEnd(currentRunId);
    } catch (error) {
      result = {
        empires: [],
        notes: [`Pörgetési hiba: ${error instanceof Error ? error.message : 'ismeretlen hiba'}`],
      };
      isLoading = false;
      isRevealing = false;
    }
  }

  function scheduleRevealEnd(runId: number): void {
    const elapsed = result.empires.length * fieldCount * 520 + 1200;
    const finishTimer = window.setTimeout(() => {
      finishReveal(runId);
    }, elapsed);
    revealTimers = [...revealTimers, finishTimer];
  }

  function finishReveal(runId: number): void {
    if (runId !== revealRunId) {
      return;
    }
    isRevealing = false;
    clearRevealTimers();
  }

  function clearRevealTimers(): void {
    revealTimers.forEach(timer => window.clearTimeout(timer));
    revealTimers = [];
  }

  async function rollTeams(): Promise<void> {
    if (hasWailsRuntime()) {
      teamRoll = await wailsRollTeams({ playerNames: playerNames() });
      return;
    }
    const shuffled = [...playerNames()];
    while (shuffled.length < 4) {
      shuffled.push(`Player ${shuffled.length + 1}`);
    }
    shuffled.sort(() => Math.random() - 0.5);
    teamRoll = { teamA: shuffled.slice(0, 2), teamB: shuffled.slice(2, 4) };
  }

  async function getOptions(selectedDLCs: string[]): Promise<OptionList> {
    if (hasWailsRuntime()) {
      return wailsGetOptionsForDLCs(selectedDLCs);
    }
    const enabled = new Set(selectedDLCs.length ? selectedDLCs : previewDlcs.map(dlc => dlc.id));
    enabled.add('base');
    return {
      dlcs: previewDlcs,
      origins: previewOrigins.filter(origin => enabled.has(origin.dlc)).map(origin => origin.name),
      species: previewSpecies.filter(item => enabled.has(item.dlc)).map(item => item.name),
    };
  }

  async function generateEmpireSet(request: GenerateRequest): Promise<GenerateResponse> {
    if (hasWailsRuntime()) {
      return wailsGenerateEmpireSet(request);
    }

    const count = mode === '2v2' ? 4 : Number(playerCount);
    const availableOrigins = origins.filter(origin => ![...teamAOriginBans, ...teamBOriginBans].includes(origin));
    const availableSpecies = species.filter(item => ![...teamASpeciesBans, ...teamBSpeciesBans].includes(item));
    const names = [...playerNames()];
    while (names.length < count) {
      names.push(`Player ${names.length + 1}`);
    }
    if (mode === '2v2') {
      names.sort(() => Math.random() - 0.5);
    }

    const normalTags = ['tech rush', 'trade league', 'relic hunter', 'ideology spread', 'storm economy'];
    const nomadTags = ['nomad arkship', 'nomad wayline logistics', 'nomad cruise fleet', 'nomad contract runner'];
    return {
      notes: ['Frontend preview mode: Wails alatt a Go backend teljes adatkeszlete porog.'],
      empires: Array.from({ length: count }, (_, index) => {
        const origin = pick(availableOrigins);
        return {
          player: names[index],
          team: mode === '2v2' ? (index < 2 ? 'Team A' : 'Team B') : 'Solo',
          origin,
          species: pick(availableSpecies),
          ethics: ['Fanatic Materialist', 'Egalitarian'],
          authority: 'Oligarchic',
          civics: ['Technocracy', 'Meritocracy'],
          traits: allowNomadicTrait ? ['Intelligent', 'Nomadic', 'Traditional'] : ['Intelligent', 'Adaptive', 'Traditional'],
          ...(request.randomizeOrigins ? {} : { origin: 'Manual Origin' }),
          ...(request.randomizeSpecies ? {} : { species: 'Manual Species' }),
          ...(request.randomizeCivics ? {} : { civics: ['Manual Civic 1', 'Manual Civic 2'] }),
          ...(request.randomizeTraits ? {} : { traits: ['Manual Traits'] }),
          homeworld: 'Continental',
          playstyleTag: origin === 'The Sacred Path' || origin === 'Forever Cruise' || origin === 'Arkship Exodus'
            ? pick(nomadTags)
            : pick(normalTags),
        };
      }),
    };
  }

  async function rerollAspect(empireIndex: number, aspect: string): Promise<void> {
    if (jokerUsedEmpireIndexes.includes(empireIndex) || isLoading || isRevealing || !fieldCanBeRerolled(aspect)) {
      return;
    }

    const key = `${empireIndex}:${aspect}`;
    jokerPendingKey = key;
    const currentEmpire = result.empires[empireIndex];
    try {
      const nextEmpire = await rerollEmpireAspect(currentEmpire, aspect);
      result = {
        ...result,
        empires: result.empires.map((empire, index) => (index === empireIndex ? nextEmpire : empire)),
      };
      jokerUsedEmpireIndexes = [...jokerUsedEmpireIndexes, empireIndex];
      jokerAnimatingKey = key;
      window.setTimeout(() => {
        if (jokerAnimatingKey === key) {
          jokerAnimatingKey = '';
        }
      }, 900);
    } finally {
      jokerPendingKey = '';
    }
  }

  async function rerollEmpireAspect(empire: Empire, aspect: string): Promise<Empire> {
    const rules = currentGenerateRequest();
    if (hasWailsRuntime()) {
      return wailsRerollEmpireAspect(main.RerollAspectRequest.createFrom({ rules, empire, aspect }));
    }

    const response = await generateEmpireSet({ ...rules, mode: 'playthrough', playerCount: 1, playerNames: [empire.player] });
    const sample = response.empires[0];
    return {
      ...empire,
      ...(aspect === 'Origin' ? { origin: sample.origin } : {}),
      ...(aspect === 'Species' ? { species: sample.species } : {}),
      ...(aspect === 'Ethics' ? { ethics: sample.ethics } : {}),
      ...(aspect === 'Authority' ? { authority: sample.authority } : {}),
      ...(aspect === 'Civics' ? { civics: sample.civics } : {}),
      ...(aspect === 'Traits' ? { traits: sample.traits } : {}),
      ...(aspect === 'Homeworld' ? { homeworld: sample.homeworld } : {}),
    };
  }

  function currentGenerateRequest(): GenerateRequest {
    return {
      mode,
      playerCount,
      playerNames: playerNames(),
      enabledDLCs,
      teamAOriginBans,
      teamBOriginBans,
      teamASpeciesBans,
      teamBSpeciesBans,
      allowCrisisBuilds,
      allowNomadicTrait,
      randomizeRulesSet: true,
      randomizeOrigins,
      randomizeSpecies,
      randomizeCivics,
      randomizeTraits,
    };
  }

  function empireFields(empire: Empire): EmpireField[] {
    return [
      { label: 'Origin', value: empire.origin },
      { label: 'Species', value: empire.species },
      { label: 'Ethics', value: empire.ethics.join(' / ') },
      { label: 'Authority', value: empire.authority },
      { label: 'Civics', value: empire.civics.join(' + ') },
      { label: 'Traits', value: empire.traits.join(', ') },
      { label: 'Homeworld', value: empire.homeworld },
    ];
  }

  function rollingText(label: string): string {
    const samples: Record<string, string[]> = {
      Origin: ['Void Dwellers', 'Clone Army', 'Under One Rule', 'Prosperous Unification'],
      Species: ['Lithoid', 'Aquatic', 'Humanoid', 'Machine'],
      Ethics: ['Fanatic Materialist', 'Xenophile', 'Gestalt Consciousness', 'Militarist'],
      Authority: ['Imperial', 'Corporate', 'Oligarchic', 'Hive Mind'],
      Civics: ['Technocracy + Meritocracy', 'Anglers + Catalytic Processing', 'Manual Civic 1 + Manual Civic 2'],
      Traits: ['Intelligent, Adaptive, Traditional', 'Manual Traits', 'Aquatic, Thrifty, Enduring'],
      Homeworld: ['Ocean', 'Continental', 'Tomb World', 'Alpine'],
    };
    const values = samples[label] ?? ['Scanning...', 'Rolling...', 'Calculating...'];
    return values[label.length % values.length];
  }

  function revealDelay(empireIndex: number, fieldIndex: number): number {
    return empireIndex * fieldCount * 520 + fieldIndex * 520;
  }

  function fieldCanBeRerolled(label: string): boolean {
    if (label === 'Origin') {
      return randomizeOrigins;
    }
    if (label === 'Species') {
      return randomizeSpecies;
    }
    if (label === 'Civics') {
      return randomizeCivics;
    }
    if (label === 'Traits') {
      return randomizeTraits;
    }
    return true;
  }

  function playerNames(): string[] {
    return playerNamesText
      .split(/\r?\n|,/)
      .map(name => name.trim())
      .filter(Boolean);
  }

  function hasWailsRuntime(): boolean {
    return typeof window !== 'undefined' && 'go' in window;
  }

  function pick(values: string[]): string {
    if (!values.length) {
      return 'Nincs elerheto opcio';
    }
    return values[Math.floor(Math.random() * values.length)];
  }

  async function toggleDLC(id: string): Promise<void> {
    if (id === 'base') {
      return;
    }
    enabledDLCs = enabledDLCs.includes(id)
      ? enabledDLCs.filter(item => item !== id)
      : [...enabledDLCs, id];
    await refreshOptions();
  }

  async function selectAllDLCs(): Promise<void> {
    enabledDLCs = dlcs.map(dlc => dlc.id);
    await refreshOptions();
  }

  async function baseOnly(): Promise<void> {
    enabledDLCs = ['base'];
    await refreshOptions();
  }

  function toggle(list: string[], value: string, limit: number): string[] {
    if (list.includes(value)) {
      return list.filter(item => item !== value);
    }
    return [...list.slice(Math.max(0, list.length - limit + 1)), value];
  }

  function switchMode(nextMode: 'playthrough' | '2v2'): void {
    mode = nextMode;
    playerCount = nextMode === '2v2' ? 4 : 1;
  }
</script>

<main class="min-h-screen overflow-x-hidden bg-zinc-950 text-zinc-100">
  <section class="mx-auto flex min-h-screen w-full max-w-[1600px] flex-col gap-6 px-6 py-6">
    <header class="app-header">
      <div>
        <p class="text-sm font-semibold uppercase tracking-wide text-cyan-300">Stellaris random empire generator</p>
        <h1 class="mt-1 text-4xl font-black text-white">2v2 es playthrough porgeto</h1>
      </div>
      <div class="flex flex-wrap gap-2">
        <button class:active-mode={mode === '2v2'} class="mode-button" on:click={() => switchMode('2v2')}>
          2v2
        </button>
        <button class:active-mode={mode === 'playthrough'} class="mode-button" on:click={() => switchMode('playthrough')}>
          Solo
        </button>
        {#if mode === 'playthrough'}
          <input
            aria-label="Player count"
            class="w-20 rounded border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
            min="1"
            max="8"
            type="number"
            bind:value={playerCount}
          />
        {/if}
      </div>
    </header>

    <div class="grid flex-1 gap-6 2xl:grid-cols-[480px_1fr]">
      <aside class="control-column">
        <section class="panel">
          <div class="flex items-center justify-between gap-3">
            <h2 class="panel-title mb-0">DLC selector</h2>
            <div class="flex gap-2">
              <button class="small-button" on:click={selectAllDLCs}>All</button>
              <button class="small-button" on:click={baseOnly}>Base</button>
            </div>
          </div>
          <div class="mt-3 grid max-h-72 gap-2 overflow-y-auto pr-1">
            {#each dlcs as dlc}
              <button
                class:selected={enabledDLCs.includes(dlc.id)}
                class:locked={dlc.id === 'base'}
                class="dlc-row"
                on:click={() => toggleDLC(dlc.id)}
              >
                <span>{dlc.name}</span>
                <small>{dlc.type}</small>
              </button>
            {/each}
          </div>
        </section>

        <section class="panel">
          <h2 class="panel-title">Jatekosok</h2>
          <textarea
            class="names-input"
            rows="5"
            bind:value={playerNamesText}
            placeholder="Nev soronkent vagy vesszovel elvalasztva"
          />
          <button class="secondary-button mt-3" on:click={rollTeams}>Teamek sorsolasa</button>
          {#if teamRoll.teamA.length || teamRoll.teamB.length}
            <div class="team-roll">
              <div>
                <p>Team A</p>
                <strong>{teamRoll.teamA.join(' + ')}</strong>
              </div>
              <div>
                <p>Team B</p>
                <strong>{teamRoll.teamB.join(' + ')}</strong>
              </div>
            </div>
          {/if}
        </section>

        <section class="panel">
          <h2 class="panel-title">Haziszabalyok</h2>
          <div class="rule-grid">
            <label class="toggle-row">
              <input type="checkbox" bind:checked={allowCrisisBuilds} />
              <span>Player lehet crisis build</span>
            </label>
            <label class="toggle-row">
              <input type="checkbox" bind:checked={allowNomadicTrait} />
              <span>Nomadic trait engedelyezve</span>
            </label>
            <label class="toggle-row">
              <input type="checkbox" bind:checked={randomizeOrigins} />
              <span>Origin random</span>
            </label>
            <label class="toggle-row">
              <input type="checkbox" bind:checked={randomizeSpecies} />
              <span>Species random</span>
            </label>
            <label class="toggle-row">
              <input type="checkbox" bind:checked={randomizeCivics} />
              <span>Civics random</span>
            </label>
            <label class="toggle-row">
              <input type="checkbox" bind:checked={randomizeTraits} />
              <span>Traits random</span>
            </label>
          </div>
        </section>

        {#if mode === '2v2'}
          <section class="ban-panel">
            <div class="panel">
              <h2 class="panel-title">Team A banok</h2>
              <div class="ban-group">
                <p>Origin ban ({teamAOriginBans.length}/2)</p>
                <div class="chip-grid">
                  {#each origins as origin}
                    <button
                      class:selected={teamAOriginBans.includes(origin)}
                      class="chip"
                      on:click={() => teamAOriginBans = toggle(teamAOriginBans, origin, 2)}
                    >
                      {origin}
                    </button>
                  {/each}
                </div>
              </div>
              <div class="ban-group">
                <p>Species ban ({teamASpeciesBans.length}/1)</p>
                <div class="chip-grid compact">
                  {#each species as item}
                    <button
                      class:selected={teamASpeciesBans.includes(item)}
                      class="chip"
                      on:click={() => teamASpeciesBans = toggle(teamASpeciesBans, item, 1)}
                    >
                      {item}
                    </button>
                  {/each}
                </div>
              </div>
            </div>

            <div class="panel">
              <h2 class="panel-title">Team B banok</h2>
              <div class="ban-group">
                <p>Origin ban ({teamBOriginBans.length}/2)</p>
                <div class="chip-grid">
                  {#each origins as origin}
                    <button
                      class:selected={teamBOriginBans.includes(origin)}
                      class="chip"
                      on:click={() => teamBOriginBans = toggle(teamBOriginBans, origin, 2)}
                    >
                      {origin}
                    </button>
                  {/each}
                </div>
              </div>
              <div class="ban-group">
                <p>Species ban ({teamBSpeciesBans.length}/1)</p>
                <div class="chip-grid compact">
                  {#each species as item}
                    <button
                      class:selected={teamBSpeciesBans.includes(item)}
                      class="chip"
                      on:click={() => teamBSpeciesBans = toggle(teamBSpeciesBans, item, 1)}
                    >
                      {item}
                    </button>
                  {/each}
                </div>
              </div>
            </div>
          </section>
        {/if}

        <button class="generate-button" disabled={isLoading} on:click={generate}>
          {isLoading ? 'Generalas...' : isRevealing ? 'Uj porgetes inditasa' : 'Empire-ek porgetese'}
        </button>
      </aside>

      <section class="result-column">
        {#if result.notes.length}
          <div class="notes">
            {#each result.notes as note}
              <span>{note}</span>
            {/each}
          </div>
        {/if}

        {#if !result.empires.length}
          <div class="empty-state">
            <p>Keszen all a porgeto.</p>
            <strong>Nyomj ra az Empire-ek porgetese gombra.</strong>
          </div>
        {/if}

        <div class="empire-grid">
          {#each result.empires as empire, empireIndex}
            <article class:revealing={isRevealing} class="empire-card">
              <div class="flex items-start justify-between gap-4">
                <div>
                  <p class="text-sm font-semibold text-cyan-300">{empire.team}</p>
                  <h2 class="text-2xl font-bold text-white">{empire.player}</h2>
                </div>
                <div class="card-status">
                  <span class="tag">{empire.playstyleTag}</span>
                  <span class:joker-used={jokerUsedEmpireIndexes.includes(empireIndex)} class="joker-status">
                    {jokerUsedEmpireIndexes.includes(empireIndex) ? 'Joker used' : 'Joker ready'}
                  </span>
                </div>
              </div>

              <dl class="mt-6 grid gap-4 md:grid-cols-2">
                {#each empireFields(empire) as field, fieldIndex}
                  {@const delay = revealDelay(empireIndex, fieldIndex)}
                  {@const jokerKey = `${empireIndex}:${field.label}`}
                  <div class:traits-field={field.label === 'Traits'} class="field-block">
                    <div class="field-header">
                      <dt>{field.label}</dt>
                      {#if fieldCanBeRerolled(field.label)}
                        <button
                          class="joker-button"
                          class:joker-pending={jokerPendingKey === jokerKey}
                          disabled={jokerUsedEmpireIndexes.includes(empireIndex) || isLoading || isRevealing || jokerPendingKey !== ''}
                          on:click={() => rerollAspect(empireIndex, field.label)}
                        >
                          {jokerPendingKey === jokerKey ? 'Pörög...' : 'Joker'}
                        </button>
                      {/if}
                    </div>
                    <dd
                      class:spin-field={isRevealing}
                      class:revealed-field={!isRevealing}
                      class:joker-highlight={jokerAnimatingKey === jokerKey}
                      style={`--reveal-delay: ${delay}ms`}
                    >
                      <span class="rolling-value">{rollingText(field.label)}</span>
                      <span class="final-value">{field.value}</span>
                    </dd>
                  </div>
                {/each}
              </dl>
            </article>
          {/each}
        </div>
      </section>
    </div>
  </section>
</main>
