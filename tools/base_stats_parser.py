#!/usr/bin/python

import csv

# Generates go/ts baes stats files from assets/db_inputs/basestats

BASE_DIR = ""

DIR_PATH = "assets/db_inputs/basestats/"
OUTPUT_PATH = "sim/core/"

BASE_MP = "octbasempbyclass.txt"
MELEE_CRIT = "chancetomeleecrit.txt"
MELEE_CRIT_BASE = "chancetomeleecritbase.txt"
SPELL_CRIT = "chancetospellcrit.txt"
SPELL_CRIT_BASE = "chancetospellcritbase.txt"
COMBAT_RATINGS = "combatratings.txt"
RATING_SCALAR = "octclasscombatratingscalar.txt"

BASE_LEVEL = 90

Offs = {
    "Warrior": 0,
    "Paladin": 1,
    "Hunter": 2,
    "Rogue": 3,
    "Priest": 4,
    "Death Knight": 5,
    "Shaman": 6,
    "Mage": 7,
    "Warlock": 8,
    "Monk": 9,
    "Druid": 10,
}

#Warrior	Paladin	Hunter	Rogue	Priest	Death Knight	Shaman	Mage	Warlock	Monk	Druid
def GenIndexedDb(file : str):
    db = {}
    with open(file) as tsv:
        first = True
        for line in csv.reader(tsv, delimiter="\t"):
            if first:
                first = False
                continue
            db[line[0]] = line[1:]
    return db

def GenRowIndexedDb(file : str):
    db = {}
    with open(file) as tsv:
        first = True
        for col in zip(*[line for line in csv.reader(tsv, delimiter='\t')]):
            if first:
                first = False
                continue
            db[col[0]] = list(col[1:])
    return db

class ClassStats:
    BaseMp : dict
    MCrit : dict
    SCrit : dict
    MCritBase : dict
    SCritBase : dict
    CombatRatings : dict

def GenExtraStatsGoFile(cs: ClassStats):
    header = '''
package core

// **************************************
// AUTO GENERATED BY BASE_STATS_PARSER.PY
// **************************************

import (
	"github.com/wowsims/mop/sim/core/proto"
	"github.com/wowsims/mop/sim/core/stats"
)

'''
    output = header
    output += f"const ExpertisePerQuarterPercentReduction = {float(cs.CombatRatings['expertise'][BASE_LEVEL-1])/4}\n"
    output += f"const HasteRatingPerHastePercent = {cs.CombatRatings['haste melee'][BASE_LEVEL-1]}\n"
    output += f"const CritRatingPerCritPercent = {cs.CombatRatings['crit melee'][BASE_LEVEL-1]}\n"
    output += f"const PhysicalHitRatingPerHitPercent = {cs.CombatRatings['hit melee'][BASE_LEVEL-1]}\n"
    output += f"const SpellHitRatingPerHitPercent = {cs.CombatRatings['hit spell'][BASE_LEVEL-1]}\n"
    output += f"const DodgeRatingPerDodgePercent = {cs.CombatRatings['dodge'][BASE_LEVEL-1]}\n"
    output += f"const ParryRatingPerParryPercent = {cs.CombatRatings['parry'][BASE_LEVEL-1]}\n"
    output += f"const MasteryRatingPerMasteryPoint = {cs.CombatRatings['mastery'][BASE_LEVEL-1]}\n"

    output += '''var CritPerAgiMaxLevel = map[proto.Class]float64{
proto.Class_ClassUnknown: 0.0,'''
    for c in ["Warrior", "Paladin", "Hunter", "Rogue", "Priest", "Death Knight", "Shaman", "Mage", "Warlock", "Monk", "Druid"]:
        cName = c.split()
        cName = ''.join(cName)
        mc = 1/float(cs.MCrit[str(BASE_LEVEL)][Offs[c]])
        output += f"\nproto.Class_Class{cName}: {mc:.8f},"
    output += "\n}\n"

    output += '''var CritPerIntMaxLevel = map[proto.Class]float64{
proto.Class_ClassUnknown: 0.0,'''
    for c in ["Warrior", "Paladin", "Hunter", "Rogue", "Priest", "Death Knight", "Shaman", "Mage", "Warlock", "Monk", "Druid"]:
        cName = c.split()
        cName = ''.join(cName)
        mc = 1/float(cs.SCrit[str(BASE_LEVEL)][Offs[c]])
        output += f"\nproto.Class_Class{cName}: {mc:.8f},"
    output += "\n}\n"

    output += '''var ExtraClassBaseStats = map[proto.Class]stats.Stats{
proto.Class_ClassUnknown: {},'''
    for c in ["Warrior", "Paladin", "Hunter", "Rogue", "Priest", "Death Knight", "Shaman", "Mage", "Warlock", "Monk", "Druid"]:
        cName = c.split()
        cName = ''.join(cName)
        output += f"\nproto.Class_Class{cName}: {{"
        mp = float(cs.BaseMp[str(BASE_LEVEL)][Offs[c]])
        scb = float(cs.SCritBase["1"][Offs[c]])*100
        mcb = float(cs.MCritBase["1"][Offs[c]])*100
        output += f"\n stats.Mana: {mp:.4f},"
        output += f"\n stats.SpellCritPercent: {scb:.4f},"
        output += f"\n stats.PhysicalCritPercent: {mcb:.4f},"
        output += "\n},"
    output += "\n}\n"
    return output


if __name__ == "__main__":
    args = ClassStats()
    args.BaseMp = GenIndexedDb(BASE_DIR + DIR_PATH + BASE_MP)
    args.MCrit = GenIndexedDb(BASE_DIR + DIR_PATH + MELEE_CRIT)
    args.SCrit = GenIndexedDb(BASE_DIR + DIR_PATH + SPELL_CRIT)
    args.MCritBase = GenIndexedDb(BASE_DIR + DIR_PATH + MELEE_CRIT_BASE)
    args.SCritBase = GenIndexedDb(BASE_DIR + DIR_PATH + SPELL_CRIT_BASE)
    args.CombatRatings = GenRowIndexedDb(BASE_DIR + DIR_PATH + COMBAT_RATINGS)

    output = GenExtraStatsGoFile(args)
    fname = BASE_DIR + OUTPUT_PATH + "base_stats_auto_gen.go"
    print(f"Writing stats to: {fname}")
    f = open(fname, "w")
    f.write(output)
    f.close()
