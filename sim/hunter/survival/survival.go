package survival

import (
	"time"

	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/proto"
	"github.com/wowsims/mop/sim/core/stats"
	"github.com/wowsims/mop/sim/hunter"
)

func RegisterSurvivalHunter() {
	core.RegisterAgentFactory(
		proto.Player_SurvivalHunter{},
		proto.Spec_SpecSurvivalHunter,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewSurvivalHunter(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_SurvivalHunter)
			if !ok {
				panic("Invalid spec value for Survival Hunter!")
			}
			player.Spec = playerSpec
		},
	)
}

func (hunter *SurvivalHunter) Initialize() {
	// Initialize global Hunter spells
	hunter.Hunter.Initialize()

	// hunter.registerExplosiveShotSpell()
	hunter.registerBlackArrowSpell()
	// Apply SV Hunter mastery
	schoolsAffectedBySurvivalMastery := []stats.SchoolIndex{
		stats.SchoolIndexNature,
		stats.SchoolIndexFire,
		stats.SchoolIndexArcane,
		stats.SchoolIndexFrost,
		stats.SchoolIndexShadow,
	}
	baseMasteryRating := hunter.GetStat(stats.MasteryRating)
	for _, school := range schoolsAffectedBySurvivalMastery {
		hunter.PseudoStats.SchoolDamageDealtMultiplier[school] *= hunter.getMasteryBonus(baseMasteryRating)
	}

	hunter.AddOnMasteryStatChanged(func(sim *core.Simulation, oldMasteryRating float64, newMasteryRating float64) {
		for _, school := range schoolsAffectedBySurvivalMastery {
			hunter.PseudoStats.SchoolDamageDealtMultiplier[school] /= hunter.getMasteryBonus(oldMasteryRating)
			hunter.PseudoStats.SchoolDamageDealtMultiplier[school] *= hunter.getMasteryBonus(newMasteryRating)
		}
	})

	hunter.applySurvivalPassives()

	// Survival Spec Bonus
	hunter.MultiplyStat(stats.Agility, 1.1)
}
func (hunter *SurvivalHunter) getMasteryBonus(masteryRating float64) float64 {
	return 1.08 + ((masteryRating / core.MasteryRatingPerMasteryPoint) * 0.01)
}

func (hunter *SurvivalHunter) applySurvivalPassives() {
	hunter.applyViperVenom()
	hunter.applyImprovedSerpentSting()
	hunter.applyTrapMastery()
	// TODO
	// hunter.applyLockAndLoad()
	// hunter.applySerpentSpread()
}

func (hunter *SurvivalHunter) applyViperVenom() {
	hunter.RegisterAura(core.Aura{
		Label:    "Viper Venom",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnPeriodicDamageDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell == hunter.SerpentSting {
				focusMetrics := hunter.NewFocusMetrics(core.ActionID{SpellID: 118974})
				hunter.AddFocus(sim, 3, focusMetrics)
			}
		},
	})
}

func (svHunter *SurvivalHunter) applyImprovedSerpentSting() {
	damageMultiplier := 1.5
	svHunter.SerpentSting.DamageMultiplierAdditive = damageMultiplier

	svHunter.ImprovedSerpentSting = svHunter.RegisterSpell(core.SpellConfig{
		ActionID:                 core.ActionID{SpellID: 82834},
		SpellSchool:              core.SpellSchoolNature,
		ProcMask:                 core.ProcMaskDirect,
		ClassSpellMask:           hunter.HunterSpellSerpentSting,
		Flags:                    core.SpellFlagPassiveSpell,
		DamageMultiplier:         1,
		DamageMultiplierAdditive: damageMultiplier,
		CritMultiplier:           svHunter.CritMultiplier(1, 0),
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := (3240.4 * 5) + 0.16*spell.RangedAttackPower()
			dmg := baseDamage * 0.15
			spell.CalcAndDealDamage(sim, target, dmg, spell.OutcomeMeleeSpecialCritOnly)
		},
	})
}

func (svHunter *SurvivalHunter) applyTrapMastery() {
	svHunter.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_Cooldown_Flat,
		ClassMask: hunter.HunterSpellBlackArrow | hunter.HunterSpellExplosiveTrap,
		TimeValue: -(time.Second * 6),
	})

	svHunter.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		ClassMask:  hunter.HunterSpellBlackArrow | hunter.HunterSpellExplosiveTrap,
		FloatValue: .30,
	})
}

func NewSurvivalHunter(character *core.Character, options *proto.Player) *SurvivalHunter {
	survivalOptions := options.GetSurvivalHunter().Options

	svHunter := &SurvivalHunter{
		Hunter: hunter.NewHunter(character, options, survivalOptions.ClassOptions),
	}

	svHunter.SurvivalOptions = survivalOptions
	// Todo: Is there a better way to do this?

	return svHunter
}

type SurvivalHunter struct {
	*hunter.Hunter
}

func (svHunter *SurvivalHunter) GetHunter() *hunter.Hunter {
	return svHunter.Hunter
}

func (svHunter *SurvivalHunter) Reset(sim *core.Simulation) {
	svHunter.Hunter.Reset(sim)
}
