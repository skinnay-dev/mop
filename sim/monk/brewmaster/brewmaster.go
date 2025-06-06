package brewmaster

import (
	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/proto"
	"github.com/wowsims/mop/sim/core/stats"
	"github.com/wowsims/mop/sim/monk"
)

func RegisterBrewmasterMonk() {
	core.RegisterAgentFactory(
		proto.Player_BrewmasterMonk{},
		proto.Spec_SpecBrewmasterMonk,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewBrewmasterMonk(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_BrewmasterMonk)
			if !ok {
				panic("Invalid spec value for Brewmaster Monk!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewBrewmasterMonk(character *core.Character, options *proto.Player) *BrewmasterMonk {
	monkOptions := options.GetBrewmasterMonk()

	bm := &BrewmasterMonk{
		Monk: monk.NewMonk(character, monkOptions.Options.ClassOptions, options.TalentsString),
	}

	bm.AddStatDependency(stats.Strength, stats.AttackPower, 1)
	bm.AddStatDependency(stats.Agility, stats.AttackPower, 2)

	// Brewmaster monks does a flat 85% of total damage as well as AP per DPS being 11 instead of 14
	bm.PseudoStats.DamageDealtMultiplier *= 0.85

	return bm
}

type BrewmasterMonk struct {
	*monk.Monk
}

func (bm *BrewmasterMonk) GetMonk() *monk.Monk {
	return bm.Monk
}

func (bm *BrewmasterMonk) Initialize() {
	bm.Monk.Initialize()
	bm.RegisterSpecializationEffects()
}

func (bm *BrewmasterMonk) ApplyTalents() {
	bm.Monk.ApplyTalents()
	bm.ApplyArmorSpecializationEffect(stats.Stamina, proto.ArmorType_ArmorTypeLeather, 120225)

	bm.registerGuard()
}

func (bm *BrewmasterMonk) Reset(sim *core.Simulation) {
	bm.Monk.Reset(sim)
}

func (bm *BrewmasterMonk) RegisterSpecializationEffects() {
	bm.RegisterMastery()
}

func (bm *BrewmasterMonk) RegisterMastery() {
}
