package hunter

import (
	"time"

	"github.com/wowsims/mop/sim/core"
)

func (hunter *Hunter) registerCobraShotSpell() {

	csMetrics := hunter.NewFocusMetrics(core.ActionID{SpellID: 77767})
	hunter.CobraShot = hunter.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 77767},
		SpellSchool:    core.SpellSchoolNature,
		ProcMask:       core.ProcMaskRangedSpecial,
		ClassSpellMask: HunterSpellCobraShot,
		Flags:          core.SpellFlagAPL,
		FocusCost: core.FocusCostOptions{
			Cost: 0,
		},
		MissileSpeed: 40,
		MinRange:     0,
		MaxRange:     40,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      time.Second,
				CastTime: time.Millisecond * 2000,
			},
			IgnoreHaste: true, // Hunter GCD is locked at 1.0s
			ModifyCast: func(_ *core.Simulation, spell *core.Spell, cast *core.Cast) {
				cast.CastTime = spell.CastTime()
			},
			CastTime: func(spell *core.Spell) time.Duration {
				ss := hunter.RangedSwingSpeed()
				return time.Duration(float64(spell.DefaultCast.CastTime) / ss)
			},
		},
		DamageMultiplier: 1,
		CritMultiplier:   hunter.DefaultCritMultiplier(),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := hunter.AutoAttacks.Ranged().CalculateNormalizedWeaponDamage(sim, spell.RangedAttackPower()) * .77
			intFocus := core.TernaryFloat64(hunter.T13_2pc.IsActive(), 14*2, 14)
			hunter.AddFocus(sim, intFocus, csMetrics)

			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedHitAndCrit)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				if hunter.SerpentSting.Dot(target).IsActive() {
					hunter.SerpentSting.Dot(target).Apply(sim) // Refresh to cause new total snapshot
				}
				spell.DealDamage(sim, result)
			})

		},
	})
}
