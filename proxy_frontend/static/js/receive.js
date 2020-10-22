var id;
var url = "";
var serverAddr = window.location.protocol + "//" + window.location.host;

var lggr;

var elDivID = document.getElementById("id");

var elButtonGetURL = document.getElementById("get-url-button");
var elAURL = document.getElementById("url-anchor")
var elDivQRCode = document.getElementById("qrcode")

var elDivLog = document.getElementById('log');

const params = getURLParams();

document.getElementById('hostname').textContent = serverAddr;

if(params.logs && (params.logs !== "false" || params.logs !== "0")) {
	lggr = new logger(elDivLog);
	elDivLog.style.display = "block";
} else {
	lggr = { log: function() {} };
}

function getParamsString(params) {
	const paramStrings = [];
	for(var key in params) {
		if(params[key]) paramStrings.push(key + '=' + params[key]);
	}
	return paramStrings.join('&');
}

function setID(newID) {
	if(id !== newID) {
		lggr.log("ID change:", id || "none", newID);
		id = newID;
		elDivID.innerText = newID;
		elButtonGetURL.disabled = false;
		elDivQRCode.src = serverAddr + '/qrcode?id=' + newID + '&url=' + serverAddr + '/send?id=' + newID
	}
}

function setURLBox(text, ref) {
	elAURL.innerHTML = text;
	if(ref) {
		elAURL.href = url;
	} else {
		elAURL.removeAttribute("href")
	}
}

var xhrURL = new XMLHttpRequest();
xhrURL.onload = function () {
	try {
		var response = JSON.parse(xhrURL.response)
		lggr.log("XHR URL Response:", response)
		if(response.url) {
			url = response.url;
			setURLBox(url, url);
		} else {
			setURLBox("Status: No URL found")
		}
		if(!id || !id.length) {
			if(response.id) {
				if(response.id.length) {
					setID(response.id);
				}
			}
		}
	} catch (error) {
		lggr.log("XHR URL Error:", error, xhrURL.response)
	}
}
xhrURL.onloadstart = function() {
	setURLBox("Status: Fetching...");
}
xhrURL.onerror = function() {
	setURLBox("Status: Error!")
}

elButtonGetURL.onclick = function() {
	xhrURL.open('GET', serverAddr + '/url/' + id);
	xhrURL.send();
}


// Registration
var interval;
var intervalID;
var xhrRegister = new XMLHttpRequest();
var registerFunc = function() {
	xhrRegister.open('GET', serverAddr + '/register');
	xhrRegister.send();
}
xhrRegister.onload = function () {
	try {
		var response = JSON.parse(xhrRegister.response)
		if(response.id) {
			setID(response.id);
			if(interval !== response.timeout * 1000) {
				interval = response.timeout * 1000;
				clearInterval(intervalID);
				intervalID = setInterval(registerFunc, interval)
			}
		}
	} catch (error) {
		lggr.log("XHR Registation Error:", error)
	}
}

registerFunc();
