# Reviewed Wyckoff benchmark

Sparse reviewed hypotheses; not investment accuracy or calibrated probabilities.
Replay feeds only OHLCV through each cutoff; labels and future outcomes never enter Analyze.
Hit = event date lies inside the reviewed interval and is emitted at some cutoff from interval end through analysis_as_of.
This measures retrospective discoverability, not earliest real-time confirmation. Recent-event pruning makes final-cutoff recall a separate observation.

| Metric | Before | After |
|---|---:|---:|
| Positive | 36 | 36 |
| Hits | 15 | 15 |
| Rejected | 20 | 20 |
| RejectedEmitted | 1 | 1 |
| Uncertain | 17 | 17 |
| Unsupported | 0 | 0 |
| StructureAmbiguous | 2 | 2 |

RejectedEmitted counts rejected proposed labels only on reviewed intervals; it is not global precision.
Uncertain, unsupported taxonomy and structure-ambiguous negatives are excluded from positive/rejected metrics.
Correct candidates are hypothesis labels, not confirmed events. Corrected replacement labels are not added as extra positives without their own correct record.

| Stock | Event / structure | Kind | Label | Before hit | After hit | First seen after | At final cutoff after |
|---|---|---|---|---|---|---|---|
| BCM | E001 / BCM_CONTEXT_202607 | uncertain |  | False | False |  | False |
| BCM | E002 / BCM_CONTEXT_202607 | positive | SC | False | False |  | False |
| BCM | E003 / BCM_CONTEXT_202607 | positive | AR | True | True | 2026-09-17 | True |
| BCM | E004 / BCM_CONTEXT_202607 | positive | ST | True | True | 2026-10-13 | True |
| BCM | E005 / BCM_CONTEXT_202607 | positive | ST | True | True | 2026-10-14 | True |
| BCM | E006 / BCM_CONTEXT_202607 | rejected | SOS | False | False |  | False |
| BCM | E007 / BCM_TR_202610 | positive | SOS | False | False |  | False |
| BCM | E008 / BCM_TR_202610 | uncertain |  | False | False |  | False |
| CTD | E001 / CTD_CONTEXT_202607 | uncertain |  | False | False |  | False |
| CTD | E002 / CTD_CONTEXT_202607 | positive | AR | False | False |  | False |
| CTD | E003 / CTD_CONTEXT_202607 | positive | ST | False | False |  | False |
| CTD | E004 / CTD_CONTEXT_202607 | positive | SOS | True | True | 2026-10-20 | False |
| CTD | E005 / CTD_CONTEXT_202607 | uncertain |  | False | False |  | False |
| CTD | E006 / CTD_CONTEXT_202607 | rejected | Spring | False | False |  | False |
| CTD | E007 / CTD_CONTEXT_202607 | structure_ambiguous | SOS | False | False |  | False |
| CTD | E008 / CTD_TR_202609 | positive | SOS | False | False |  | False |
| GEX | E001 / GEX_CONTEXT_202607 | uncertain |  | False | False |  | False |
| GEX | E002 / GEX_CONTEXT_202607 | uncertain |  | False | False |  | False |
| GEX | E003 / GEX_CONTEXT_202607 | positive | AR | True | True | 2026-09-15 | True |
| GEX | E004 / GEX_CONTEXT_202607 | positive | ST | True | True | 2026-10-09 | True |
| GEX | E005 / GEX_CONTEXT_202607 | rejected | Spring | False | False |  | False |
| GEX | E006 / GEX_TR_202610 | rejected | AR | False | False |  | False |
| GEX | E007 / GEX_CONTEXT_202607 | rejected | SOS | False | False |  | False |
| GEX | E008 / GEX_TR_202610 | rejected | LPS | False | False |  | False |
| HDG | E001 / HDG_TR_202607 | uncertain |  | False | False |  | False |
| HDG | E002 / HDG_TR_202607 | positive | AR | False | False |  | False |
| HDG | E003 / HDG_TR_202607 | positive | ST | False | False |  | False |
| HDG | E004 / HDG_TR_202607 | positive | SOS | False | False |  | False |
| HDG | E005 / HDG_TR_202607 | rejected | SOW | True | True | 2026-10-05 | False |
| HDG | E006 / HDG_TR_202610 | uncertain |  | False | False |  | False |
| HDG | E007 / HDG_TR_202610 | positive | AR | False | False |  | False |
| HDG | E008 / HDG_TR_202610 | uncertain |  | False | False |  | False |
| HDG | E009 / HDG_TR_202607 | rejected | Spring | False | False |  | False |
| HDG | E010 / HDG_TR_202610 | uncertain |  | False | False |  | False |
| LPB | E001 / LPB_TR_202606 | positive | BC | False | False |  | False |
| LPB | E002 / LPB_TR_202606 | positive | AR | False | False |  | False |
| LPB | E003 / LPB_TR_202606 | positive | ST | False | False |  | False |
| LPB | E004 / LPB_TR_202606 | positive | SOW | False | False |  | False |
| LPB | E005 / LPB_TR_202606 | positive | LPSY | False | False |  | False |
| LPB | E006 / LPB_MARKDOWN_202609 | rejected | SC | False | False |  | False |
| LPB | E007 / LPB_MARKDOWN_202609 | rejected | SOS | False | False |  | False |
| LPB | E008 / LPB_MARKDOWN_202609 | rejected | SC | False | False |  | False |
| MSB | E001 / MSB_TR_202607 | positive | Spring | False | False |  | False |
| MSB | E002 / MSB_TR_202607 | uncertain |  | False | False |  | False |
| MSB | E003 / MSB_TR_202607 | positive | SOS | True | True | 2026-09-23 | False |
| MSB | E004 / MSB_TR_202607 | positive | LPS | False | False |  | False |
| MSB | E005 / MSB_TR_202607 | rejected | SC | False | False |  | False |
| MSB | E006 / MSB_TR_202607 | rejected | SOW | False | False |  | False |
| MSB | E007 / MSB_TR_202607 | rejected | LPS | False | False |  | False |
| NLG | E001 / NLG_TR_202609_SHORT | uncertain |  | False | False |  | False |
| NLG | E002 / NLG_TR_202609_SHORT | positive | AR | False | False |  | False |
| NLG | E003 / NLG_TR_202609_SHORT | positive | Spring | False | False |  | False |
| NLG | E004 / NLG_TR_202609_SHORT | uncertain |  | False | False |  | False |
| NLG | E005 / NLG_TR_202609_SHORT | positive | SOS | False | False |  | False |
| NLG | E006 / NLG_TR_202609_SHORT | uncertain |  | False | False |  | False |
| NLG | E007 / NLG_TR_202609_SHORT | rejected | Test | False | False |  | False |
| NLG | E008 / NLG_CONTEXT_202607 | structure_ambiguous | SOS | False | False |  | False |
| NLG | E009 / NLG_CONTEXT_202607 | positive | SC | True | True | 2026-07-28 | False |
| NLG | E010 / NLG_CONTEXT_202607 | uncertain |  | False | False |  | False |
| NLG | E011 / NLG_CONTEXT_202607 | positive | ST | True | True | 2026-10-15 | True |
| NVL | E001 / NVL_TR_202609 | uncertain |  | False | False |  | False |
| NVL | E002 / NVL_TR_202609 | positive | SC | True | True | 2026-10-02 | True |
| NVL | E003 / NVL_TR_202609 | positive | AR | True | True | 2026-10-13 | True |
| NVL | E004 / NVL_TR_202609 | positive | ST | False | False |  | False |
| NVL | E005 / NVL_TR_202609 | rejected | SOS | False | False |  | False |
| NVL | E006 / NVL_TR_202609 | rejected | Spring | False | False |  | False |
| NVL | E007 / NVL_TR_202609 | rejected | LPS | False | False |  | False |
| SZC | E001 / SZC_TR_202607 | positive | PS | True | True | 2026-09-21 | True |
| SZC | E002 / SZC_TR_202607 | positive | SC | True | True | 2026-09-21 | True |
| SZC | E003 / SZC_TR_202607 | positive | AR | True | True | 2026-09-21 | True |
| SZC | E004 / SZC_TR_202607 | positive | ST | False | False |  | False |
| SZC | E005 / SZC_TR_202607 | positive | SOS | True | True | 2026-10-23 | False |
| SZC | E006 / SZC_TR_202607 | uncertain |  | False | False |  | False |
| SZC | E007 / SZC_TR_202607 | rejected | Spring | False | False |  | False |
| SZC | E008 / SZC_TR_202607 | rejected | SC | False | False |  | False |

Baseline analyzer: `9bec474abd22eefd1b20efa4d548dff0dea9ecdc`. Candidate comparison run: https://github.com/trungvdn/investment-ai/actions/runs/41449686493 .

Observed recall of reviewed positive hypotheses is 15/36 (41.7%) before and after. The Spring fix repairs an impossible predicate, but shows no gain on these nine samples. HDG E005 is the one rejected SOW emitted during replay. MSB and NLG Spring labels remain missed. The analyzer lacks structure identifiers and explicit reaccumulation/redistribution output; nested-range attribution is not scored as if supported.
