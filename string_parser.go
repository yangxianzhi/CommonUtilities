package commonutilities

import "fmt"

type StringParser struct {
	buffer        string
	curLineNumber int
	startIndex    int
	endIndex      int
}

func New(inString string) *StringParser {
	inLen := len(inString)
	var startIndex, endIndex = -1, -1
	if inLen > 0 {
		startIndex = 0
		endIndex = inLen
	}
	return &StringParser{buffer: inString, curLineNumber: 1, startIndex: startIndex, endIndex: endIndex}
}

//GetBuffer:
//Returns a pointer to the string object
func (s *StringParser) GetStream() string { return s.buffer }

func (s *StringParser) ParserIsEmpty() bool {
	if len(s.buffer) == 0 {
		return true
	}

	if -1 == s.startIndex || -1 == s.endIndex {
		return true
	}

	if s.startIndex > s.endIndex {
		panic(fmt.Sprintf("s.startIndex:%d > s.endIndex:%d", s.startIndex, s.endIndex))
	}

	return false // parser ok to parse
}

func (s *StringParser) ConsumeWord() string {
	return s.ConsumeUntil(sNonWordMask)
}

// ConsumeWhitespace
//Keeps on going until non-whitespace
func (s *StringParser) ConsumeWhitespace(){
	s.ConsumeUntil(sWhitespaceMask)
}

// ConsumeUntilWhitespace
//+ rt 8.19.99
//returns whatever is avaliable until non-whitespace
func (s *StringParser) ConsumeUntilWhitespace() string {
	return s.ConsumeUntil(sEOLWhitespaceMask)
}

func (s *StringParser) ConsumeUntilDigit() string {
	return s.ConsumeUntil(sDigitMask)
}

//Returns the current character, doesn't move past it.
func (s *StringParser) PeekFast() byte {
	if s.startIndex != -1 {
		return s.buffer[s.startIndex]
	} else {
		return 0
	}
}

func (s *StringParser) GetCurrentPosition() int {
	return s.startIndex
}

func (s *StringParser) GetCurrentLineNumber() int {
	return s.curLineNumber
}

// ConsumeUntilStop
//Returns all the data before inStopChar
func (s *StringParser) ConsumeUntilStop(inStop byte) string {
	if s.ParserIsEmpty() {
		return ""
	}

	originalStartIndex := s.startIndex
	for (s.startIndex < s.endIndex) && (s.buffer[s.startIndex] != inStop) {
		s.advanceMark()
	}
	return s.buffer[originalStartIndex:s.startIndex]
}

// ConsumeUntil
//Assumes 'inMask' is a 255-char array of booleans. Set this array
//to a mask of what the stop characters are. true means stop character.
//You may also pass in one of the many prepackaged masks defined above.
func (s *StringParser) ConsumeUntil(inMask []uint8) string {
	if s.ParserIsEmpty() {
		return ""
	}

	originalStartIndex := s.startIndex
	for (s.startIndex < s.endIndex) && !(inMask[uint8(s.buffer[s.startIndex])] == 1) {
		s.advanceMark()
	}
	return s.buffer[originalStartIndex:s.startIndex]
}

func (s *StringParser) ConsumeLength(inLength int) string {
	if s.ParserIsEmpty(){
		return ""
	}
	//sanity check to make sure we aren't being told to run off the end of the
	//buffer
	if(s.endIndex - s.startIndex) < inLength {
		inLength = s.endIndex - s.startIndex
	}
	ret := s.buffer[s.startIndex:s.startIndex+inLength]

	if inLength>0{
		for i:=0; i< inLength; i++{
			s.advanceMark()
		}
	} else {
		s.startIndex += inLength // ***may mess up line number if we back up too much
	}

	return  ret
}

// ConsumeInteger
// Returns whatever integer is currently in the stream
func (s *StringParser) ConsumeInteger() ( outString string, theValue uint8) {
	if s.ParserIsEmpty(){
		return
	}

	originalStartIndex := s.startIndex
	for (s.startIndex < s.endIndex) && (s.buffer[s.startIndex] >= '0') && (s.buffer[s.startIndex] <= '9') {
		theValue = (theValue * 10 ) + (s.buffer[s.startIndex] - '0')
		s.advanceMark()
	}
	outString = s.buffer[originalStartIndex:s.startIndex]
	return
}

func (s *StringParser) ConsumeFloat() ( theFloat float32) {
	if s.ParserIsEmpty(){
		return
	}
	for (s.startIndex < s.endIndex) && (s.buffer[s.startIndex] >= '0') && (s.buffer[s.startIndex] <= '9') {
		theFloat = (theFloat * 10) + float32(s.buffer[s.startIndex] - '0')
		s.advanceMark()
	}
	if (s.startIndex < s.endIndex) && s.GetCurrentPosition() == '.' {
		s.advanceMark()
	}

	multiplier := float32(.1)
	for (s.startIndex < s.endIndex) && (s.buffer[s.startIndex] >= '0') && (s.buffer[s.startIndex] <= '9')   {
		theFloat += float32(multiplier * float32(s.buffer[s.startIndex] - '0'))
		multiplier *= float32(.1)
		s.advanceMark()
	}
	return
}

func (s *StringParser) ConsumeNPT() (theFloat float32) {
	if s.ParserIsEmpty(){
		return
	}
	valArray := [4]float32{0,0,0,0}
	divArray := [4]float32{1,1,1,1}
	valType,index := 0,0
	for index = 0; index < 4 ; index++ {
		for (s.startIndex < s.endIndex) && (s.buffer[s.startIndex] >= '0') && (s.buffer[s.startIndex] <= '9') {
			valArray[index] = (valArray[index] * 10) + float32(s.buffer[s.startIndex] - '0')
			divArray[index] *= 10
			s.advanceMark()
		}
		if s.startIndex >= s.endIndex || valType == 0 && index >=1 {
			break
		}
		if s.buffer[s.startIndex] == '.' && valType == 0 && index == 0 {

		} else if s.buffer[s.startIndex] == ':' && index < 2 {
			valType = 1
		} else if s.buffer[s.startIndex] == '.' && index == 2 {

		} else {
			break
		}
		s.advanceMark()
	}
	if valType == 0 {
		theFloat = valArray[0] + (valArray[1] / divArray[1])
	} else {
		theFloat = (valArray[0] * 3600) + (valArray[1] * 60) + valArray[2] + (valArray[3] / divArray[3])
	}
	return
}

func (s *StringParser) Expect(stopChar byte) bool {
	if s.ParserIsEmpty() {
		return false
	}
	if s.startIndex > s.endIndex {
		return false
	}

	if s.buffer[s.startIndex] != stopChar {
		return false
	} else {
		s.advanceMark()
		return true
	}
}

func (s *StringParser) ExpectEOL() bool {
	if s.ParserIsEmpty() {
		return false
	}

	//This function processes all legal forms of HTTP / RTSP eols.
	//They are: \r (alone), \n (alone), \r\n
	retVal := false
	if (s.startIndex < s.endIndex) && ((s.buffer[s.startIndex] == '\r') || (s.buffer[s.startIndex] == '\n')) {
		retVal = true
		s.advanceMark()
		//check for a \r\n, which is the most common EOL sequence.
		if (s.startIndex < s.endIndex) && (s.buffer[s.startIndex - 1] == '\r') && (s.buffer[s.startIndex] == '\n') {
			s.advanceMark()
		}
	}
	return retVal
}

func (s *StringParser) ConsumeEOL() (outString string) {
	if s.ParserIsEmpty() {
		return
	}

	//This function processes all legal forms of HTTP / RTSP eols.
	//They are: \r (alone), \n (alone), \r\n
	originalStartIndex := s.startIndex
	if (s.startIndex < s.endIndex) && ((s.buffer[s.startIndex] == '\r') || (s.buffer[s.startIndex] == '\n')) {
		s.advanceMark()
		//check for a \r\n, which is the most common EOL sequence.
		if (s.startIndex < s.endIndex) && (s.buffer[s.startIndex - 1] == '\r') && (s.buffer[s.startIndex] == '\n') {
			s.advanceMark()
		}
	}
	outString = s.buffer[originalStartIndex:s.startIndex]
	return
}

//GetThru:
//Works very similar to ConsumeUntil except that it moves past the stop token,
//and if it can't find the stop token it returns false
func (s *StringParser) GetThru(stopChar byte) (outString string, outBool bool) {
	outString = s.ConsumeUntilStop(stopChar)
	outBool = s.Expect(stopChar)
	return
}

//GetThruEOL:
func (s *StringParser) GetThruEOL() (outString string, outBool bool) {
	outString = s.ConsumeUntil(sEOLMask)
	outBool = s.ExpectEOL()
	return
}

// UnQuote　去掉字符串中的引号
// If a string is contained within double or single quotes
// then UnQuote() will remove them. - [sfu]
func (s *StringParser) UnQuote(inString string) string {
	// sanity check
	if len(inString) < 2 { return inString }

	// remove begining quote if it's there.
	if inString[0] == '"' || inString[0] == '\'' {
		inString = inString[1:]
	}

	// remove ending quote if it's there.
	if inString[len(inString) - 1] == '"' || inString[len(inString) - 1] == '\'' {
		inString = inString[0:len(inString) - 1]
	}
	return inString
}

//Returns some info about the stream
func (s *StringParser) GetDataParsedLen() (theValue int) {
	theValue = s.startIndex
	if theValue < 0 {
		panic(fmt.Sprintf("s.startIndex = %d < 0",theValue))
	}
	return
}

func (s *StringParser) GetDataReceivedLen() (theValue int) {
	theValue = len(s.buffer)
	if theValue < 0 {
		panic(fmt.Sprintf("len(s.buffer) = %d < 0",theValue))
	}
	return
}

func (s *StringParser) GetDataRemaining() (theValue int) {
	theValue = s.endIndex - s.startIndex
	if theValue < 0 {
		panic(fmt.Sprintf("s.endIndex - s.startIndex = %d < 0",theValue))
	}
	return
}

func (s *StringParser) advanceMark() {
	if s.ParserIsEmpty() {
		return
	}

	if (s.buffer[s.startIndex] == '\n') || ((s.buffer[s.startIndex] == '\r') && (s.buffer[s.startIndex+1] != '\n')) {
		// we are progressing beyond a line boundary (don't count \r\n twice)
		s.curLineNumber++
	}
	s.startIndex++
}
