package paladin

import (
	"time"

	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/proto"
)

func (paladin *Paladin) registerWordOfGlory() {
	isProt := paladin.Spec == proto.Spec_SpecProtectionPaladin
	actionID := core.ActionID{SpellID: 85673}

	scalingCoef := 4.84999990463
	variance := 0.1080000028
	spCoef := 0.49000000954

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ProcMask:       core.ProcMaskSpellHealing,
		SpellSchool:    core.SpellSchoolHoly,
		ClassSpellMask: SpellMaskWordOfGlory,
		MetricSplits:   4,

		MaxRange: 40,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.TernaryDuration(isProt, 0, core.GCDDefault),
				NonEmpty: isProt,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Millisecond * 1500,
			},
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				paladin.DynamicHolyPowerSpent = paladin.SpendableHolyPower()
				spell.SetMetricsSplit(paladin.DynamicHolyPowerSpent)
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return paladin.HolyPower.CanSpend(1)
		},

		DamageMultiplier: 1,
		CritMultiplier:   paladin.DefaultCritMultiplier(),
		ThreatMultiplier: 1,

		BonusCoefficient: spCoef,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseHealing := paladin.CalcAndRollDamageRange(sim, scalingCoef, variance)

			damageMultiplier := spell.DamageMultiplier
			spell.DamageMultiplier *= float64(paladin.DynamicHolyPowerSpent)
			result := spell.CalcHealing(sim, target, baseHealing, spell.OutcomeHealingCrit)
			spell.DamageMultiplier = damageMultiplier

			paladin.HolyPower.SpendUpTo(paladin.DynamicHolyPowerSpent, actionID, sim)

			spell.DealOutcome(sim, result)
		},
	})
}
