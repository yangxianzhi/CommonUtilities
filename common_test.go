package commonutilities

import (
	"regexp"
	"strings"
	"testing"
)

var (
	optionsRequest = "OPTIONS rtsp://172.22.0.172/123.ts/?channel=1&token=888888 RTSP/1.0\r\n" +
		"CSeq: 1\r\n" +
		"User-Agent: LibVLC/2.1.2 (Dor Streaming Media v1.0.0.3))\r\n\r\n"

	descriptionRequest = "DESCRIBE rtsp://192.168.1.103/live1.264?channel=1&token=888888 RTSP/1.0\r\n" +
		"CSeq: 2\r\n" +
		"User-Agent: LibVLC/2.1.5 (Dor Streaming Media v1.0.0.3))\r\n" +
		"Accept: application/sdp\r\n\r\n"

	setupRequest = "SETUP rtsp://192.168.1.105:8554/test.264/track1?channel=1&token=888888 RTSP/1.0\r\n" +
		"CSeq: 3\r\n" +
		"User-Agent: dorsvr (Dor Streaming Media v1.0.0.3)\r\n" +
		"Transport: RTP/AVP;unicast;client_port=37175-37176\r\n\r\n"

	playRequest = "PLAY rtsp://192.168.1.105:8554/test.264?channel=1&token=888888 RTSP/1.0\r\n" +
		"CSeq: 4\r\n" +
		"User-Agent: dorsvr (Dor Streaming Media v1.0.0.3)\r\n" +
		"Session: E1155C20\r\n" +
		"Range: npt=0.000-\r\n"

	teardownRequest = "TEARDOWN rtsp://192.168.1.105:8554/test.264?channel=1&token=888888 RTSP/1.0\r\n" +
		"CSeq: 5\r\n" +
		"Session: E1155C20\r\n" +
		"User-Agent: VLC media player (Dor Streaming Media v1.0.0.3))"

	announceRequest = "ANNOUNCE rtsp://192.168.199.136:8554/asdf?channel=1&token=888888 RTSP/1.0\r\n" +
		"Content-Type: application/sdp\r\n" +
		"CSeq: 2\r\n" +
		"User-Agent: Lavf57.83.100\r\n" +
		"Content-Length: 339\r\n" +
		"\r\n" +
		"v=0\r\n" +
		"o=- 0 0 IN IP4 127.0.0.1\r\n" +
		"s=RTSP Session\r\n" +
		"c=IN IP4 192.168.199.136\r\n" +
		"t=0 0\r\n" +
		"a=tool:libavformat 57.83.100\r\n" +
		"m=video 0 RTP/AVP 96\r\n" +
		"a=rtpmap:96 H264/90000\r\n" +
		"a=fmtp:96 packetization-mode=1; sprop-parameter-sets=Z0IAH52oFAFum4CAgIE=,aM48gA==; profile-level-id=42001F\r\n" +
		"a=control:streamid=0\r\n" +
		"m=audio 0 RTP/AVP 8\r\n" +
		"b=AS:64\r\n" +
		"a=control:streamid=1\r\n" +
		"\r\n"

	string1 = "RTSP 200 OK\r\nContent-Type: MeowMix\r\n\t   \n3450"
)

func TestParseRTSPRequestString(t *testing.T) {
	var tests = []struct {
		input string
		want  string
		reqURI string
	}{
		{optionsRequest, "OPTIONS", "/123.ts/"},
		{descriptionRequest, "DESCRIBE","/live1.264"},
		{setupRequest, "SETUP","/test.264/track1"},
		{teardownRequest, "TEARDOWN","/test.264"},
		{playRequest,"PLAY","/test.264"},
		{announceRequest, "ANNOUNCE","/asdf"},
	}
	for _, test := range tests {
		s := New(test.input)
		parseFirstLine(s,t,test)
		parseHeaders(s,t,test)
	}
}

func parseFirstLine(s *StringParser, t *testing.T, test struct {input string;	want  string;	reqURI string})  {
	if got := s.ConsumeWord(); got != test.want {
		t.Errorf("ConsumeWord(%q) = %v", test.input, got)
	}
	s.ConsumeWhitespace()
	if s.PeekFast() != 'r' {
		t.Errorf("GetCurrentPosition(%q) = %c", test.input, 'r')
	}
	{
		strRegex := "^((ws|wss|rtsp|rtmp|http|sip)?://)" +
			"?(([0-9a-z_!~*'().&=+$%-]+: )?[0-9a-z_!~*'().&=+$%-]+@)?" + // ftp的user@
			"(([0-9]{1,3}\\.){3}[0-9]{1,3}" + // IP形式的URL- 199.194.52.184
			"|" + // 允许IP和DOMAIN（域名）
			"([0-9a-z_!~*'()-]+\\.)*" + // 域名- www.
			"([0-9a-z][0-9a-z-]{0,61})?[0-9a-z]\\." + // 二级域名
			"[a-z]{2,6})" + // first level domain- .com or .museum
			"(:[0-9]{1,5})?" + // 端口1~5位- :8080
			"((/?)|" + // a slash isn't required if there is no file name
			"(/[0-9a-z_!~*'().;?:@&=+$,%#-]+)+/?)$"
		url := s.ConsumeUntilWhitespace()
		if matched,err := regexp.MatchString(strRegex,url) ; matched != true || err != nil {
			t.Errorf("ConsumeUntilWhitespace(%q) = %v", test.input, url)
		}

		absParser:= New(url)
		absUrl := absParser.ConsumeUntil(sURLStopConditions)
		if matched,err := regexp.MatchString(strRegex,absUrl) ; matched != true || err != nil {
			t.Errorf("ConsumeUntil(%q) = %v", url, absUrl)
		}

		urlParser := New(absUrl)
		if absUrl[0] != '/' && absUrl[0] != '*' {
			if rtsp := urlParser.ConsumeLength(7); (rtsp != "rtsp://") && (rtsp != "RTSP://") {
				t.Errorf("ConsumeLength(%q) = %v", absUrl, rtsp)
			}
			strReg := "^(([0-9]{1,3}\\.){3}[0-9]{1,3}" + // IP形式的URL- 199.194.52.184
				"|" + // 允许IP和DOMAIN（域名）
				"([0-9a-z_!~*'()-]+\\.)*" + // 域名- www.
				"([0-9a-z][0-9a-z-]{0,61})?[0-9a-z]\\." + // 二级域名
				"[a-z]{2,6})" + // first level domain- .com or .museum
				"(:[0-9]{1,5})?" // 端口1~5位- :8080
			host := urlParser.ConsumeUntilStop('/')
			if matched,err := regexp.MatchString(strReg,host); matched != true || err != nil {
				t.Errorf("ConsumeLength(%q) = %v", absUrl, host)
			}
		}
		if  reqUrl := urlParser.ConsumeLength(urlParser.GetDataRemaining()); reqUrl != test.reqURI{
			t.Errorf("ConsumeLength(%q) %q = %v", test.input , absUrl, reqUrl)
		}

		if absParser.GetDataRemaining() > 0 {
			if absParser.PeekFast() == '?' {
				if got:=absParser.ConsumeLength(1);got!="?"{
					t.Errorf("ConsumeLength(%q) = %v", test.input ,  got)
				}
				if got:=absParser.ConsumeUntilWhitespace(); got!="channel=1&token=888888" {
					t.Errorf("ConsumeUntilWhitespace(%q) = %v", test.input ,  got)
				}
			}
		}
	}
	s.ConsumeWhitespace()
	if got:= s.ConsumeUntil(sEOLMask);got != "RTSP/1.0"{
		t.Errorf("ConsumeUntil(%q) = %v", test.input ,  got)
	}
	if !s.ExpectEOL() {
		t.Errorf("ExpectEOL(%q) = %v", test.input ,  "\r\n")
	}
}

func parseHeaders(parser *StringParser, t *testing.T, test struct{input string; want string; reqURI string}){
	var keyWord string
	var isOk bool
	for parser.PeekFast() != '\r' && parser.PeekFast() != '\n' {
		if keyWord,isOk = parser.GetThru(':'); isOk {
			keyWord = strings.TrimSpace(keyWord)
			theHeaderVal := parser.ConsumeUntil(sEOLMask)

			var theEOL string
			if parser.PeekFast() == '\r' || parser.PeekFast() == '\n' {
				isOk = true
				theEOL = parser.ConsumeEOL()
			} else {
				isOk = false
			}
			for parser.PeekFast() == ' ' || parser.PeekFast() == '\t' {
				theHeaderVal += theEOL
				temp := parser.ConsumeUntil(sEOLMask)
				theHeaderVal += temp

				if parser.PeekFast() == '\r' || parser.PeekFast() == '\n' {
					isOk = true
					theEOL = parser.ConsumeEOL()
				} else {
					isOk = false
				}
			}
			theHeaderVal = strings.TrimSpace(theHeaderVal)
		} else {
			t.Errorf("GetThru(%q) = %v", test.input, keyWord)
		}
	}
}

func Test(t *testing.T) {
	s := New(string1)
	if strGot, got := s.ConsumeInteger(); got != 0 {
		t.Errorf("GetCurrentPosition(%q) = %v, %c", string1, strGot, got)
	}

	if got := s.ConsumeWord(); got != "RTSP" {
		t.Errorf("ConsumeWord(%q) = %v", string1, got)
	}
	s.ConsumeWhitespace()
	if strGot, got := s.ConsumeInteger(); got != 200 {
		t.Errorf("GetCurrentPosition(%q) = %v, %v", string1, strGot, got)
	}
	s.ConsumeWhitespace()
	if got := s.ConsumeWord(); got != "OK" {
		t.Errorf("ConsumeWord(%q) = %v", string1, got)
	}
	if got := s.ConsumeEOL(); got != "\r\n" {
		t.Errorf("ConsumeWord(%q) = %v", string1, got)
	}
}

// reverse reverses a slice of ints in place.
func reverse(s []int) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
