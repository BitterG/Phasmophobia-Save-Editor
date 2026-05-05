export namespace main {
	
	export class SaveData {
	    rebirthLevel: number;
	    level: number;
	    gold: number;
	
	    static createFrom(source: any = {}) {
	        return new SaveData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rebirthLevel = source["rebirthLevel"];
	        this.level = source["level"];
	        this.gold = source["gold"];
	    }
	}

}

