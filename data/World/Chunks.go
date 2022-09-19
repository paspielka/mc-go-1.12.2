package World

import (
	"bytes"
	"fmt"
	"github.com/edouard127/mc-go-1.12.2/save"
	"math/bits"
	"strings"

	"github.com/edouard127/mc-go-1.12.2/data/World/block"
	"github.com/edouard127/mc-go-1.12.2/nbt"
)

type ChunkPos [2]int32

type Chunk struct {
	Sections    []Section
	HeightMaps  HeightMaps
	BlockEntity []BlockEntity
	Status      ChunkStatus
}

func EmptyChunk(secs int) *Chunk {
	sections := make([]Section, secs)
	for i := range sections {
		sections[i] = Section{
			BlockCount: 0,
			States:     NewStatesPaletteContainer(16*16*16, 0),
			Biomes:     NewBiomesPaletteContainer(4*4*4, 0),
		}
	}
	return &Chunk{
		Sections: sections,
		HeightMaps: HeightMaps{
			MotionBlocking: NewBitStorage(bits.Len(uint(secs)*16), 16*16, nil),
		},
		Status: StatusEmpty,
	}
}

var biomesIDs map[string]BiomesState

var biomesNames = []string{
	"the_void",
	"plains",
	"sunflower_plains",
	"snowy_plains",
	"ice_spikes",
	"desert",
	"swamp",
	"mangrove_swamp",
	"forest",
	"flower_forest",
	"birch_forest",
	"dark_forest",
	"old_growth_birch_forest",
	"old_growth_pine_taiga",
	"old_growth_spruce_taiga",
	"taiga",
	"snowy_taiga",
	"savanna",
	"savanna_plateau",
	"windswept_hills",
	"windswept_gravelly_hills",
	"windswept_forest",
	"windswept_savanna",
	"jungle",
	"sparse_jungle",
	"bamboo_jungle",
	"badlands",
	"eroded_badlands",
	"wooded_badlands",
	"meadow",
	"grove",
	"snowy_slopes",
	"frozen_peaks",
	"jagged_peaks",
	"stony_peaks",
	"river",
	"frozen_river",
	"beach",
	"snowy_beach",
	"stony_shore",
	"warm_ocean",
	"lukewarm_ocean",
	"deep_lukewarm_ocean",
	"ocean",
	"deep_ocean",
	"cold_ocean",
	"deep_cold_ocean",
	"frozen_ocean",
	"deep_frozen_ocean",
	"mushroom_fields",
	"dripstone_caves",
	"lush_caves",
	"deep_dark",
	"nether_wastes",
	"warped_forest",
	"crimson_forest",
	"soul_sand_valley",
	"basalt_deltas",
	"the_end",
	"end_highlands",
	"end_midlands",
	"small_end_islands",
	"end_barrens",
}

func init() {
	biomesIDs = make(map[string]BiomesState, len(biomesNames))
	for i, v := range biomesNames {
		biomesIDs[v] = BiomesState(i)
	}
}

// ChunkFromSave convert save.Chunk to level.Chunk.
func ChunkFromSave(c *save.Chunk) (*Chunk, error) {
	secs := len(c.Sections)
	sections := make([]Section, secs)
	for _, v := range c.Sections {
		i := int32(v.Y) - c.YPos
		if i < 0 || i >= int32(secs) {
			return nil, fmt.Errorf("section Y value %d out of bounds", v.Y)
		}
		var err error
		sections[i].BlockCount, sections[i].States, err = readStatesPalette(v.BlockStates.Palette, v.BlockStates.Data)
		if err != nil {
			return nil, err
		}
		sections[i].Biomes, err = readBiomesPalette(v.Biomes.Palette, v.Biomes.Data)
		if err != nil {
			return nil, err
		}
		sections[i].SkyLight = v.SkyLight
		sections[i].BlockLight = v.BlockLight
	}

	motionBlocking := c.Heightmaps.MotionBlocking
	motionBlockingNoLeaves := c.Heightmaps.MotionBlockingNoLeaves
	oceanFloor := c.Heightmaps.OceanFloor
	worldSurface := c.Heightmaps.WorldSurface

	bitsForHeight := bits.Len( /* chunk height in blocks */ uint(secs) * 16)
	return &Chunk{
		Sections: sections,
		HeightMaps: HeightMaps{
			MotionBlocking:         NewBitStorage(bitsForHeight, 16*16, motionBlocking),
			MotionBlockingNoLeaves: NewBitStorage(bitsForHeight, 16*16, motionBlockingNoLeaves),
			OceanFloor:             NewBitStorage(bitsForHeight, 16*16, oceanFloor),
			WorldSurface:           NewBitStorage(bitsForHeight, 16*16, worldSurface),
		},
		Status: ChunkStatus(c.Status),
	}, nil
}

func readStatesPalette(palette []save.BlockState, data []uint64) (blockCount int16, paletteData *PaletteContainer[BlocksState], err error) {
	statePalette := make([]BlocksState, len(palette))
	for i, v := range palette {
		b, ok := block.FromID[v.Name]
		if !ok {
			return 0, nil, fmt.Errorf("unknown block id: %v", v.Name)
		}
		if v.Properties.Data != nil {
			if err := v.Properties.Unmarshal(&b); err != nil {
				return 0, nil, fmt.Errorf("unmarshal block properties fail: %v", err)
			}
		}
		s, ok := block.ToStateID[b]
		if !ok {
			return 0, nil, fmt.Errorf("unknown block: %v", b)
		}
		if !block.IsAir(s) {
			blockCount++
		}
		statePalette[i] = s
	}
	paletteData = NewStatesPaletteContainerWithData(16*16*16, data, statePalette)
	return
}

func readBiomesPalette(palette []string, data []uint64) (*PaletteContainer[BiomesState], error) {
	biomesRawPalette := make([]BiomesState, len(palette))
	var ok bool
	for i, v := range palette {
		biomesRawPalette[i], ok = biomesIDs[strings.TrimPrefix(v, "minecraft:")]
		if !ok {
			return nil, fmt.Errorf("unknown biomes: %s", v)
		}
	}
	return NewBiomesPaletteContainerWithData(4*4*4, data, biomesRawPalette), nil
}

// ChunkToSave convert level.Chunk to save.Chunk
func ChunkToSave(c *Chunk, dst *save.Chunk) (err error) {
	secs := len(c.Sections)
	sections := make([]save.Section, secs)
	for i, v := range c.Sections {
		s := &sections[i]
		states := &s.BlockStates
		biomes := &s.Biomes
		s.Y = int8(int32(i) + dst.YPos)
		states.Palette, states.Data, err = writeStatesPalette(v.States)
		if err != nil {
			return
		}
		biomes.Palette, biomes.Data = writeBiomesPalette(v.Biomes)
		s.SkyLight = v.SkyLight
		s.BlockLight = v.BlockLight
	}
	dst.Sections = sections
	dst.Heightmaps.MotionBlocking = c.HeightMaps.MotionBlocking.Raw()
	dst.Status = string(c.Status)
	return
}

func writeStatesPalette(paletteData *PaletteContainer[BlocksState]) (palette []save.BlockState, data []uint64, err error) {
	rawPalette := paletteData.palette.export()
	palette = make([]save.BlockState, len(rawPalette))
	var buffer bytes.Buffer
	for i, v := range rawPalette {
		b := block.StateList[v]
		palette[i].Name = b.ID()

		buffer.Reset()
		err = nbt.NewEncoder(&buffer).Encode(b, "")
		if err != nil {
			return
		}
		_, err = nbt.NewDecoder(&buffer).Decode(&palette[i].Properties)
		if err != nil {
			return
		}
	}
	data = append(data, paletteData.data.Raw()...)

	return
}

func writeBiomesPalette(paletteData *PaletteContainer[BiomesState]) (palette []string, data []uint64) {
	rawPalette := paletteData.palette.export()
	palette = make([]string, len(rawPalette))
	for i, v := range rawPalette {
		palette[i] = biomesNames[v]
	}
	data = append(data, paletteData.data.Raw()...)

	return
}

type HeightMaps struct {
	MotionBlocking         *BitStorage
	MotionBlockingNoLeaves *BitStorage
	OceanFloor             *BitStorage
	WorldSurface           *BitStorage
}

type BlockEntity struct {
	XZ   int8
	Y    int16
	Type int32
	Data nbt.RawMessage
}

func (b BlockEntity) UnpackXZ() (X, Z int) {
	return int((uint8(b.XZ) >> 4) & 0xF), int(uint8(b.XZ) & 0xF)
}

type Section struct {
	BlockCount int16
	States     *PaletteContainer[BlocksState]
	Biomes     *PaletteContainer[BiomesState]
	// Half a byte per light value.
	// Could be nil if not exist
	SkyLight   []byte // len() == 2048
	BlockLight []byte // len() == 2048
}

func (s *Section) GetBlock(i int) BlocksState {
	return s.States.Get(i)
}
func (s *Section) SetBlock(i int, v BlocksState) {
	if block.IsAir(s.States.Get(i)) {
		s.BlockCount--
	}
	if v != 0 {
		s.BlockCount++
	}
	s.States.Set(i, v)
}

type lightData struct {
	SkyLightMask   []byte
	BlockLightMask []byte
	SkyLight       []byte
	BlockLight     []byte
}

func bitSetRev(set []byte) []byte {
	rev := make([]byte, len(set))
	for i := range rev {
		rev[i] = ^set[i]
	}
	return rev
}
