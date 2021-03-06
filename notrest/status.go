/*
Copyright (c) 2017 Simon Schmidt

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/


package notrest

var statusCodes = map[int]string{
	100:"Continue",
	101:"Switching Protocols",
	102:"Processing",
	// -----------------------
	200:"OK",
	201:"Created",
	202:"Accepted",
	203:"Non-Authoritative Information",
	204:"No Content",
	205:"Reset Content",
	206:"Partial Content",
	207:"Multi-Status",
	208:"Already Reported",
	226:"IM Used",
	
	220:"IM Used",
	// -----------------------
	300:"Multiple Choices",
	301:"Moved Permanently",
	302:"Found (Moved Temporarily)",
	303:"See Other",
	304:"Not Modified",
	305:"Use Proxy",
	306:"(reserviert)",
	307:"Temporary Redirect",
	308:"Permanent Redirect",
	// -----------------------
	400:"Bad Request",
	401:"Unauthorized",
	402:"Payment Required",
	403:"Forbidden",
	404:"Not Found",
	405:"Method Not Allowed",
	406:"Not Acceptable",
	407:"Proxy Authentication Required",
	408:"Request Time-out",
	409:"Conflict",
	410:"Gone",
	411:"Length Required",
	412:"Precondition Failed",
	413:"Request Entity Too Large",
	414:"Request-URL Too Long",
	415:"Unsupported Media Type",
	416:"Requested range not satisfiable",
	417:"Expectation Failed",
	420:"Policy Not Fulfilled",
	421:"Misdirected Request",
	422:"Unprocessable Entity",
	423:"Locked",
	424:"Failed Dependency",
	426:"Upgrade Required",
	428:"Precondition Required",
	429:"Too Many Requests",
	431:"Request Header Fields Too Large",
	451:"Unavailable For Legal Reasons",
	
	430:"Request Header Fields Too Large",
	450:"Unavailable For Legal Reasons",
	
	418:"I'm a teapot",
	425:"Unordered Collection",
	444:"No Response",
	449:"The request should be retried after doing the appropriate action",
	499:"Client Closed Request",
	// -----------------------
	500:"Internal Server Error",
	501:"Not Implemented",
	502:"Bad Gateway",
	503:"Service Unavailable",
	504:"Gateway Time-out",
	505:"HTTP Version not supported",
	506:"Variant Also Negotiates",
	507:"Insufficient Storage",
	508:"Loop Detected",
	509:"Bandwidth Limit Exceeded",
	510:"Not Extended",
	511:"Network Authentication Required",
	// -----------------------
}
func (r *Response) Status(code int) {
	reason,ok := statusCodes[code]
	if !ok { reason,ok = statusCodes[code - (code%10 )] }
	if !ok { reason,ok = statusCodes[code - (code%100)] }
	r.code   = code
	r.reason = r.cbstr(reason)
}

