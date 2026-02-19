export namespace main {
	
	export class ScanRequest {
	    inputFilePath: string;
	    concurrency: number;
	    timeoutSeconds: number;
	    followRedirect: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ScanRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inputFilePath = source["inputFilePath"];
	        this.concurrency = source["concurrency"];
	        this.timeoutSeconds = source["timeoutSeconds"];
	        this.followRedirect = source["followRedirect"];
	    }
	}
	export class ScanRow {
	    url: string;
	    title: string;
	    components: string[];
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new ScanRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.title = source["title"];
	        this.components = source["components"];
	        this.error = source["error"];
	    }
	}
	export class ScanResponse {
	    reportPath: string;
	    total200Lines: number;
	    totalUrls: number;
	    succeeded: number;
	    failed: number;
	    rows: ScanRow[];
	
	    static createFrom(source: any = {}) {
	        return new ScanResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.reportPath = source["reportPath"];
	        this.total200Lines = source["total200Lines"];
	        this.totalUrls = source["totalUrls"];
	        this.succeeded = source["succeeded"];
	        this.failed = source["failed"];
	        this.rows = this.convertValues(source["rows"], ScanRow);
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

}

