export namespace internal {
	
	export class Node {
	    id: string;
	    name: string;
	    ip: string;
	    port: number;
	    protocol: string;
	    isMaster: boolean;
	    username?: string;
	    authMethod?: string;
	    keyPath?: string;
	
	    static createFrom(source: any = {}) {
	        return new Node(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.ip = source["ip"];
	        this.port = source["port"];
	        this.protocol = source["protocol"];
	        this.isMaster = source["isMaster"];
	        this.username = source["username"];
	        this.authMethod = source["authMethod"];
	        this.keyPath = source["keyPath"];
	    }
	}
	export class NodeStatus {
	    latency: number;
	    // Go type: time
	    lastChecked: any;
	    status: string;
	    errorMsg: string;
	
	    static createFrom(source: any = {}) {
	        return new NodeStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.latency = source["latency"];
	        this.lastChecked = this.convertValues(source["lastChecked"], null);
	        this.status = source["status"];
	        this.errorMsg = source["errorMsg"];
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
	export class TopologyData {
	    master?: Node;
	    slaves: Node[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new TopologyData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.master = this.convertValues(source["master"], Node);
	        this.slaves = this.convertValues(source["slaves"], Node);
	        this.total = source["total"];
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

