export namespace main {
	
	export class DLCOption {
	    id: string;
	    name: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new DLCOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	    }
	}
	export class Empire {
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
	
	    static createFrom(source: any = {}) {
	        return new Empire(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.player = source["player"];
	        this.team = source["team"];
	        this.origin = source["origin"];
	        this.species = source["species"];
	        this.ethics = source["ethics"];
	        this.authority = source["authority"];
	        this.civics = source["civics"];
	        this.traits = source["traits"];
	        this.homeworld = source["homeworld"];
	        this.playstyleTag = source["playstyleTag"];
	    }
	}
	export class GenerateRequest {
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
	
	    static createFrom(source: any = {}) {
	        return new GenerateRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.playerCount = source["playerCount"];
	        this.playerNames = source["playerNames"];
	        this.enabledDLCs = source["enabledDLCs"];
	        this.teamAOriginBans = source["teamAOriginBans"];
	        this.teamBOriginBans = source["teamBOriginBans"];
	        this.teamASpeciesBans = source["teamASpeciesBans"];
	        this.teamBSpeciesBans = source["teamBSpeciesBans"];
	        this.allowCrisisBuilds = source["allowCrisisBuilds"];
	        this.allowNomadicTrait = source["allowNomadicTrait"];
	        this.randomizeRulesSet = source["randomizeRulesSet"];
	        this.randomizeOrigins = source["randomizeOrigins"];
	        this.randomizeSpecies = source["randomizeSpecies"];
	        this.randomizeCivics = source["randomizeCivics"];
	        this.randomizeTraits = source["randomizeTraits"];
	    }
	}
	export class GenerateResponse {
	    empires: Empire[];
	    notes: string[];
	
	    static createFrom(source: any = {}) {
	        return new GenerateResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.empires = this.convertValues(source["empires"], Empire);
	        this.notes = source["notes"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class OptionList {
	    dlcs: DLCOption[];
	    origins: string[];
	    species: string[];
	
	    static createFrom(source: any = {}) {
	        return new OptionList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dlcs = this.convertValues(source["dlcs"], DLCOption);
	        this.origins = source["origins"];
	        this.species = source["species"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RerollAspectRequest {
	    rules: GenerateRequest;
	    empire: Empire;
	    aspect: string;
	
	    static createFrom(source: any = {}) {
	        return new RerollAspectRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rules = this.convertValues(source["rules"], GenerateRequest);
	        this.empire = this.convertValues(source["empire"], Empire);
	        this.aspect = source["aspect"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TeamRoll {
	    teamA: string[];
	    teamB: string[];
	
	    static createFrom(source: any = {}) {
	        return new TeamRoll(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.teamA = source["teamA"];
	        this.teamB = source["teamB"];
	    }
	}
	export class TeamRollRequest {
	    playerNames: string[];
	
	    static createFrom(source: any = {}) {
	        return new TeamRollRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.playerNames = source["playerNames"];
	    }
	}

}

