package runtime

type MemoryManager struct {
	Default  string
	memories map[string]int
}

func NewMemoryManager(defaultMem string) *MemoryManager {
	memories := map[string]int{
		"128MB": 128,
		"256MB": 256,
		"512MB": 512,
		"1GB":   1024,
		"2GB":   2048,
		"4GB":   4096,
		"8GB":   8192,
	}
	return &MemoryManager{
		Default:  defaultMem,
		memories: memories,
	}
}

func (m *MemoryManager) Get(memory string) int {
	mem, exists := m.memories[memory]
	if !exists {
		return m.memories[m.Default]
	}
	return mem
}

func (m *MemoryManager) List() map[string]int {
	return m.memories
}
