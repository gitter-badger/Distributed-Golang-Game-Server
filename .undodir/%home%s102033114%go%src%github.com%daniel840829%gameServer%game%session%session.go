Vim�UnDo� ��Pq����A������8%c�C9)�U�ϴ  �                                  [�jx    _�                        '    ����                                                                                                                                                                                                                                                                                                                                                             [�jb    �        �      /	//. "github.com/daniel840829/gameServer2/uuid"�        �      +	"github.com/daniel840829/gameServer2/user"�        �      ,	. "github.com/daniel840829/gameServer2/msg"5�_�                    �       ����                                                                                                                                                                                                                                                                                                                                                             [�jt    �   �   �  �      Ifun: (sb *SessionStateBase) SetStateCode(code SessionInfo_SessionState) {5�_�                     �       ����                                                                                                                                                                                                                                                                                                                                                             [�jw    �              �   package session       import (   /	//"github.com/daniel840829/gameServer2/entity"   +	. "github.com/daniel840829/gameServer/msg"   *	"github.com/daniel840829/gameServer/user"   .	//. "github.com/daniel840829/gameServer/uuid"   !	log "github.com/sirupsen/logrus"   "	"google.golang.org/grpc/metadata"   
	"strconv"   	"sync"   	"time"   )       type MsgChannel struct {   	DataCh     chan (interface{})   	StopSignal chan (struct{})   }       func (m *MsgChannel) Close() {   		select {   	case <-m.StopSignal:   		return   		default:   		close(m.StopSignal)   	}   }       4func NewMsgChannel(bufferNumber int32) *MsgChannel {   	return &MsgChannel{   5		DataCh:     make(chan (interface{}), bufferNumber),   '		StopSignal: make(chan (struct{}), 1),   	}   }       0func NewMsgChannelManager() *MsgChannelManager {   	return &MsgChannelManager{   		make(map[string]*MsgChannel),   	}   }       type MsgChannelManager struct {   	c map[string]*MsgChannel   }       Nfunc (m *MsgChannelManager) AddMsgChan(name string, bufferNumber int32) bool {   	if _, ok := m.c[name]; ok {   		return false   	}   (	m.c[name] = NewMsgChannel(bufferNumber)   	return true   }       Afunc (m *MsgChannelManager) GetMsgChan(name string) *MsgChannel {   	return m.c[name]   }       7func (m *MsgChannelManager) CloseMsgChan(name string) {   	if ch, ok := m.c[name]; ok {   		ch.Close()   		delete(m.c, name)   	}   }       type sessionManager struct {   	Sessions map[int64]*Session   	sync.RWMutex   }       @func (sm *sessionManager) MakeSession(info *SessionInfo) int64 {   	s := NewSession(info)   
	sm.Lock()   	sm.Sessions[s.Info.Uuid] = s   	sm.Unlock()   	return s.Info.Uuid   }       Ufunc (sm *sessionManager) CreateSessionFromAgent(sessionInfo *SessionInfo) *Session {   W	characterInfo := sessionInfo.UserInfo.OwnCharacter[sessionInfo.UserInfo.UsedCharacter]   	log.Debug(characterInfo)   .	s := sm.Sessions[sm.MakeSession(sessionInfo)]   		return s   }       ?func (sm *sessionManager) GetSession(md metadata.MD) *Session {   	mdid := md.Get("session-id")   	if len(mdid) == 0 {   		return nil   	}   -	id, err := strconv.ParseInt(mdid[0], 10, 64)   	if err != nil {   		return nil   	}   	s, ok := sm.Sessions[id]   		if !ok {   
		return s   	}   
	s.RLock()   	if s.User != nil {   		uname := md.Get("uname")   		if len(uname) == 0 {   			s.RUnlock()   			return nil   2		} else if s.User.UserInfo.UserName != uname[0] {   			s.RUnlock()   			return nil   		}   	}   	s.RUnlock()   		return s   }       -func NewSession(info *SessionInfo) *Session {   	s := &Session{   		Info:              info,   ,		MsgChannelManager: NewMsgChannelManager(),   #		PlayerInfo:        &PlayerInfo{},   	}   ^	for i := int32(SessionInfo_NoSession); i <= int32(SessionInfo_GameServerWaitReconnect); i++ {   L		ss := SessionStateFactory.makeSessionState(s, SessionInfo_SessionState(i))   !		s.States = append(s.States, ss)   	}   '	s.SetState(int32(SessionInfo_OnStart))   	s.State.CreateSession()   	s.AddMsgChan("GameFrame", 5)   		return s   }       type Session struct {   	Info       *SessionInfo   	State      SessionState   	SessionKey int64   	User       *user.User   	States     []SessionState   	Room       *Room   	sync.RWMutex   	PlayerInfo *PlayerInfo   	TeamNo     int32   	InputPool  *MsgChannel   	InputStamp int64   	*MsgChannelManager   }       /func (s *Session) GetPlayerInfo() *PlayerInfo {   	if s.User == nil {   		return nil   	}   1	s.PlayerInfo.UserName = s.User.UserInfo.UserName   +	s.PlayerInfo.UserId = s.User.UserInfo.Uuid   /	if s.User.UserInfo.UsedCharacter == int64(0) {   3		for id, c := range s.User.UserInfo.OwnCharacter {   %			s.User.UserInfo.UsedCharacter = id   			s.PlayerInfo.Character = c   			break   		}   		} else {   V		s.PlayerInfo.Character = s.User.UserInfo.OwnCharacter[s.User.UserInfo.UsedCharacter]   	}   	s.PlayerInfo.TeamNo = s.TeamNo   	return s.PlayerInfo   }       /func (s *Session) SetState(state_index int32) {    	s.State = s.States[state_index]   }       type SessionState interface {   	SetSession(s *Session) bool   '	SetStateCode(SessionInfo_SessionState)   (	GetStateCode() SessionInfo_SessionState   	CreateSession() int64   ,	Login(uname string, pswd string) *user.User   	Logout() bool   7	Regist(uname string, pswd string, info ...string) bool   &	CreateRoom(setting *RoomSetting) bool   	EnterRoom(roomId int64) bool   	DeleteRoom() bool   	ReadyRoom() bool   	LeaveRoom() bool   	StartRoom() bool   )	SettingCharacter(*CharacterSetting) bool   	SettingRoom() bool   	CancelReady() bool   	EndRoom() bool   	String() string   	Lock()   	HandleInput(input *Input)   		Unlock()   	StartGame()   	Reconnect()   	EndConnection()   	WaitReconnect()   	End()   }       9func (sb *SessionStateBase) SetSession(s *Session) bool {   	if sb.Session != nil {   		return false   	}   	sb.Session = s   	return true   }       #func (sb *SessionStateBase) End() {   	log.Debug("origin")   }       -func (sb *SessionStateBase) String() string {   :	return SessionInfo_SessionState_name[int32(sb.StateCode)]   }       )func (sb *SessionStateBase) StartGame() {   	//TODO   }       )func (sb *SessionStateBase) Reconnect() {       }       -func (sb *SessionStateBase) WaitReconnect() {       }   -func (sb *SessionStateBase) EndConnection() {       }       Ifunc (sb *SessionStateBase) SetStateCode(code SessionInfo_SessionState) {   	sb.StateCode = code   }   Efunc (sb *SessionStateBase) GetStateCode() SessionInfo_SessionState {   	return sb.StateCode   }   3func (sb *SessionStateBase) CreateSession() int64 {   		return 0   }       7func (sb *SessionStateBase) HandleInput(input *Input) {   }       Ifunc (sb *SessionStateBase) Login(uname string, pswd string) *user.User {   	return nil   }   +func (sb *SessionStateBase) Logout() bool {   	return false   }   Tfunc (sb *SessionStateBase) Regist(uname string, pswd string, info ...string) bool {   	return false   }   Cfunc (sb *SessionStateBase) CreateRoom(setting *RoomSetting) bool {   	return false   }   :func (sb *SessionStateBase) EnterRoom(roomId int64) bool {   	return false   }   /func (sb *SessionStateBase) DeleteRoom() bool {   	return false   }   .func (sb *SessionStateBase) ReadyRoom() bool {   	return false   }   .func (sb *SessionStateBase) LeaveRoom() bool {   	return false   }   .func (sb *SessionStateBase) StartRoom() bool {   	return false   }   Ffunc (sb *SessionStateBase) SettingCharacter(*CharacterSetting) bool {   	return false   }   0func (sb *SessionStateBase) SettingRoom() bool {   	return false   }   ,func (sb *SessionStateBase) EndRoom() bool {   	return false   }       0func (sb *SessionStateBase) CancelReady() bool {   	return false   }       type SessionStateBase struct {   #	StateCode SessionInfo_SessionState   	Session   *Session   	sync.RWMutex   }       %type SessionStateGameOnStart struct {   	SessionStateBase   }       3func (ssgos *SessionStateGameOnStart) StartGame() {   	//firsttimeGetGameframe   +	ssgos.Session.Room.GameStart <- struct{}{}   3	ssgos.Session.SetState(int32(SessionInfo_Playing))   E	log.Debug("state code :", int32(ssgos.Session.State.GetStateCode()))   }       !type SessionStatePlaying struct {   	SessionStateBase   }       :func (sb *SessionStatePlaying) HandleInput(input *Input) {   -	if sb.Session.InputStamp > input.TimeStamp {   )		log.Debug("input sequence is disorder")   	}   %	sb.Session.InputPool.DataCh <- input   }       0func (sb *SessionStatePlaying) WaitReconnect() {   @	sb.Session.SetState(int32(SessionInfo_GameServerWaitReconnect))   }       0func (sb *SessionStatePlaying) EndConnection() {   1	sb.Session.SetState(int32(SessionInfo_GameOver))   B	log.Debug("state code :", int32(sb.Session.State.GetStateCode()))   	sb.Session.State.End()   }       -type SessionStateWaitingReconnection struct {   	SessionStateBase   }       8func (sswr *SessionStateWaitingReconnection) Waiting() {   	go func() {   #		c := time.After(10 * time.Second)   	END:   		for {   			select {   			case <-c:   				break END   			}   		}       2		log.Warn("Exceed waiting time, end connection!")   		sswr.EndConnection()   	}()   }       >func (sswr *SessionStateWaitingReconnection) EndConnection() {   3	sswr.Session.SetState(int32(SessionInfo_GameOver))   D	log.Debug("state code :", int32(sswr.Session.State.GetStateCode()))   	sswr.Session.State.End()   }       :func (sswr *SessionStateWaitingReconnection) StartGame() {   	sswr.Reconnect()   D	log.Debug("state code :", int32(sswr.Session.State.GetStateCode()))   }       :func (sswr *SessionStateWaitingReconnection) Reconnect() {   2	sswr.Session.SetState(int32(SessionInfo_Playing))   D	log.Debug("state code :", int32(sswr.Session.State.GetStateCode()))   4	sswr.Session.Room.Client[sswr.Session] = struct{}{}   }       "type SessionStateGameOver struct {   	SessionStateBase   }       )func (ssgo *SessionStateGameOver) End() {   	ssgo.Session.Lock()   	if ssgo.Session.Room == nil {   		ssgo.Session.Unlock()   		return   	}   ,	ssgo.Session.Room.PlayerLeave(ssgo.Session)   2	log.Debug("ssgo.Session.Room", ssgo.Session.Room)   	ssgo.Session.Room = nil   	ssgo.Session.Unlock()   }       !type sessionStateFactory struct {   }       ufunc (sf *sessionStateFactory) makeSessionState(session *Session, state_code SessionInfo_SessionState) SessionState {   	var s SessionState   	switch state_code {   	case SessionInfo_OnStart:    		s = &SessionStateGameOnStart{}   	case SessionInfo_Playing:   		s = &SessionStatePlaying{}   	case SessionInfo_GameOver:   		s = &SessionStateGameOver{}   *	case SessionInfo_GameServerWaitReconnect:   (		s = &SessionStateWaitingReconnection{}   		default:   		s = &SessionStateBase{}   	}   		s.Lock()   	s.SetSession(session)   	s.SetStateCode(state_code)   	s.Unlock()   		return s   }       var Manager *sessionManager       ,var SessionStateFactory *sessionStateFactory       func init() {   	Manager = &sessionManager{   %		Sessions: make(map[int64]*Session),   	}   -	SessionStateFactory = &sessionStateFactory{}   }5��