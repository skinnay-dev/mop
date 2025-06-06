package combat

import "github.com/wowsims/mop/sim/core"

func (comRogue *CombatRogue) applyCombatPotency() {
	if comRogue.Talents.CombatPotency == 0 {
		return
	}

	const procChance = 0.2
	energyBonus := 5.0 * float64(comRogue.Talents.CombatPotency)
	energyMetrics := comRogue.NewEnergyMetrics(core.ActionID{SpellID: 35546})

	comRogue.RegisterAura(core.Aura{
		Label:    "Combat Potency",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Landed() && (spell.ProcMask.Matches(core.ProcMaskMeleeOHAuto) || spell.SpellID == 86392) { // 86392 = Main Gauche
				if sim.RandomFloat("Combat Potency") < procChance {
					comRogue.AddEnergy(sim, energyBonus, energyMetrics)
				}
			}
		},
	})
}
