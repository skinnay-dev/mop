package retribution

import (
	"math"
	"time"

	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/paladin"
)

func (ret *RetributionPaladin) registerInquisition() {
	actionID := core.ActionID{SpellID: 84963}
	inquisitionDuration := time.Second * 20

	inquisitionAura := ret.RegisterAura(core.Aura{
		Label:    "Inquisition" + ret.Label,
		ActionID: actionID,
		Duration: inquisitionDuration,
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.3,
		School:     core.SpellSchoolHoly,
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: 10,
	})

	// Inquisition self-buff.
	ret.Inquisition = ret.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		Flags:          core.SpellFlagAPL,
		ProcMask:       core.ProcMaskEmpty,
		SpellSchool:    core.SpellSchoolHoly,
		ClassSpellMask: paladin.SpellMaskInquisition,
		MetricSplits:   4,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				ret.DynamicHolyPowerSpent = ret.SpendableHolyPower()
				spell.SetMetricsSplit(ret.DynamicHolyPowerSpent)
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return ret.HolyPower.CanSpend(1)
		},

		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			duration := inquisitionDuration * time.Duration(ret.DynamicHolyPowerSpent+core.TernaryInt32(ret.T11Ret4pc.IsActive(), 1, 0))

			// Inquisition behaves like a dot with DOT_REFRESH, which means you'll never lose your current tick
			if spell.RelatedSelfBuff.IsActive() {
				carryover := spell.RelatedSelfBuff.RemainingDuration(sim).Seconds()
				result := math.Floor(carryover / 2)
				carryover -= result * 2
				duration += core.DurationFromSeconds(carryover)
			}

			spell.RelatedSelfBuff.Duration = duration
			spell.RelatedSelfBuff.Activate(sim)

			ret.HolyPower.SpendUpTo(ret.DynamicHolyPowerSpent, actionID, sim)
		},

		RelatedSelfBuff: inquisitionAura,
	})
}
