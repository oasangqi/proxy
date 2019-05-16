package conf

import (
	"github.com/BurntSushi/toml"
	xtime "github.com/oasangqi/proxy/pkg/time"
	"time"
)

type Config struct {
	Interval  xtime.Duration
	PidFile   string
	WebSocket *WebsocketOptions `toml:"websocket"`
	Round     *RoundOptions     `toml:"buffer"`
	Redis     *Redis            `toml:"redis"`
	Servers   map[string]Server `toml:"servers"`
	Http      *HttpOptions
}

type WebsocketOptions struct {
	Addrs            []string
	TlsOpen          bool
	TlsAddrs         []string
	CertFile         string
	PrivateFile      string
	HandShakeTimeOut xtime.Duration
	Debug            bool
}

type RoundOptions struct {
	Timer        int
	TimerSize    int
	Reader       int
	ReadBuf      int
	ReadBufSize  int
	Writer       int
	WriteBuf     int
	WriteBufSize int
}

type HttpOptions struct {
	Addr   string
	LogDir string
	Open   bool
}

type Redis struct {
	Network      string
	Addr         string
	Auth         string
	Active       int
	Idle         int
	DialTimeout  xtime.Duration
	ReadTimeout  xtime.Duration
	WriteTimeout xtime.Duration
	IdleTimeout  xtime.Duration
}

type Server struct {
	Vid  int64
	Name string
	Addr string
}

const filePath = "proxy.toml"

var Conf *Config

func Default() *Config {
	return &Config{
		Interval: xtime.Duration(time.Second * 10),
		PidFile:  "proxy.pid",
		WebSocket: &WebsocketOptions{
			TlsOpen:          false,
			HandShakeTimeOut: xtime.Duration(time.Second * 10),
			Debug:            false,
		},
		Round: &RoundOptions{
			Timer:        4,
			TimerSize:    1024,
			Reader:       4,
			ReadBuf:      1024,
			ReadBufSize:  16 * 1024,
			Writer:       4,
			WriteBuf:     1024,
			WriteBufSize: 16 * 1024,
		},
		Servers: make(map[string]Server, 0),
		Http: &HttpOptions{
			Addr:   ":8000",
			LogDir: "logs",
			Open:   false,
		},
	}
}

func (c *Config) Load() error {
	if _, err := toml.DecodeFile(filePath, c); err != nil {
		return err
	}
	return nil
}

func Init() (err error) {
	Conf = Default()
	err = Conf.Load()
	return err
}
