export namespace core {
	
	export class ContactRecord {
	    id: string;
	    category: string;
	    fields: string[];
	    // Go type: time
	    updated_at: any;
	    deleted: boolean;
	    synced: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ContactRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.category = source["category"];
	        this.fields = source["fields"];
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.deleted = source["deleted"];
	        this.synced = source["synced"];
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
	export class ExportPayload {
	    filename: string;
	    headers: string[];
	    rows: string[][];
	
	    static createFrom(source: any = {}) {
	        return new ExportPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filename = source["filename"];
	        this.headers = source["headers"];
	        this.rows = source["rows"];
	    }
	}
	export class SyncStatus {
	    is_syncing: boolean;
	    last_sync: string;
	    error: string;
	    details: string;
	
	    static createFrom(source: any = {}) {
	        return new SyncStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.is_syncing = source["is_syncing"];
	        this.last_sync = source["last_sync"];
	        this.error = source["error"];
	        this.details = source["details"];
	    }
	}
	export class UpdateCheckResult {
	    has_update: boolean;
	    latest_ver: string;
	    release_notes: string;
	    download_url: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.has_update = source["has_update"];
	        this.latest_ver = source["latest_ver"];
	        this.release_notes = source["release_notes"];
	        this.download_url = source["download_url"];
	    }
	}

}

