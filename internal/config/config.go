package config

type Config struct {
	TargetDir  string
	SourceFile string
	Workers    int
	Rewrite    bool
	WorkTime   int
	SleepTime  int
}
