# Structure-scoped Wyckoff benchmark

scope-v1: boundary distance <=0.05 reviewed width each (0.05-0.15 held as LOW_CONFIDENCE); formation overlap >=0.5 shorter interval; cost separation >=0.05; no fallback for missing IDs

Evaluation only: annotations never enter Analyze. Automatic IDs are mapped by frozen boundaries and formation time, not by event-label agreement.
Unattributed, unmapped, low-confidence and ambiguous predictions are not counted as correct or rejected-target emissions. All reviewed positive labels remain in the denominator.
Reviewed candidates are hypotheses; sparse labels cannot establish global precision or investment accuracy.

| Metric | Baseline | Candidate |
|---|---:|---:|
| Positive | 36 | 36 |
| Hits | 6 | 6 |
| Rejected | 22 | 22 |
| RejectedEmitted | 0 | 0 |
| Uncertain | 17 | 17 |
| MappedPredictions | 14 | 14 |
| UnattributedPredictions | 103 | 103 |
| UnmappedPredictions | 55 | 55 |
| AmbiguousPredictions | 0 | 0 |
| LowConfidencePredictions | 7 | 7 |

Unlike the legacy score, all reviewed rejections are eligible by their own structure, including same-date contradictory labels on nested ranges. No missing-ID fallback is used.
Prediction counts are unique ticker / predicted structure / event type / date over prefix replay within the review window. They are not false-positive counts.

| Stock | Reviewed structure | Mapping coverage | Positive hits / total | Rejected emitted / total | Uncertain |
|---|---|---|---:|---:|---:|
| BCM | BCM_CONTEXT_202607 | True | 0/4 | 0/1 | 1 |
| BCM | BCM_TR_202610 | False | 0/1 | 0/0 | 1 |
| CTD | CTD_CONTEXT_202607 | True | 1/3 | 0/2 | 2 |
| CTD | CTD_TR_202609 | True | 1/1 | 0/0 | 0 |
| GEX | GEX_CONTEXT_202607 | True | 0/2 | 0/2 | 2 |
| GEX | GEX_TR_202610 | False | 0/0 | 0/2 | 0 |
| HDG | HDG_TR_202607 | True | 1/3 | 0/2 | 1 |
| HDG | HDG_TR_202610 | False | 0/1 | 0/0 | 3 |
| LPB | LPB_MARKDOWN_202609 | TARGET_UNDEFINED | 0/0 | 0/3 | 0 |
| LPB | LPB_TR_202606 | True | 0/5 | 0/0 | 0 |
| MSB | MSB_TR_202607 | True | 1/3 | 0/3 | 1 |
| NLG | NLG_CONTEXT_202607 | True | 0/2 | 0/1 | 1 |
| NLG | NLG_TR_202609_SHORT | True | 2/3 | 0/1 | 3 |
| NVL | NVL_TR_202609 | True | 0/3 | 0/3 | 1 |
| SZC | SZC_TR_202607 | False | 0/5 | 0/2 | 1 |

| Stock | Event | Reviewed structure | Kind | Label | Scoped hit | Disposition | First seen |
|---|---|---|---|---|---|---|---|
| BCM | E001 | BCM_CONTEXT_202607 | uncertain |  | False | EXCLUDED_UNCERTAIN |  |
| BCM | E002 | BCM_CONTEXT_202607 | positive | SC | False | NO_PREDICTION |  |
| BCM | E003 | BCM_CONTEXT_202607 | positive | AR | False | UNATTRIBUTED |  |
| BCM | E004 | BCM_CONTEXT_202607 | positive | ST | False | UNATTRIBUTED |  |
| BCM | E005 | BCM_CONTEXT_202607 | positive | ST | False | UNATTRIBUTED |  |
| BCM | E006 | BCM_CONTEXT_202607 | rejected | SOS | False | NO_PREDICTION |  |
| BCM | E007 | BCM_TR_202610 | positive | SOS | False | NO_PREDICTION |  |
| BCM | E008 | BCM_TR_202610 | uncertain |  | False | EXCLUDED_UNCERTAIN |  |
| CTD | E001 | CTD_CONTEXT_202607 | uncertain |  | False | EXCLUDED_UNCERTAIN |  |
| CTD | E002 | CTD_CONTEXT_202607 | positive | AR | False | NO_PREDICTION |  |
| CTD | E003 | CTD_CONTEXT_202607 | positive | ST | False | NO_PREDICTION |  |
| CTD | E004 | CTD_CONTEXT_202607 | positive | SOS | True | MATCHED_TARGET | 2026-10-20 |
| CTD | E005 | CTD_CONTEXT_202607 | uncertain |  | False | EXCLUDED_UNCERTAIN |  |
| CTD | E006 | CTD_CONTEXT_202607 | rejected | Spring | False | LOW_CONFIDENCE |  |
| CTD | E007 | CTD_CONTEXT_202607 | rejected | SOS | False | OTHER_STRUCTURE |  |
| CTD | E008 | CTD_TR_202609 | positive | SOS | True | MATCHED_TARGET | 2026-10-07 |
| GEX | E001 | GEX_CONTEXT_202607 | uncertain |  | False | EXCLUDED_UNCERTAIN |  |
| GEX | E002 | GEX_CONTEXT_202607 | uncertain |  | False | EXCLUDED_UNCERTAIN |  |
| GEX | E003 | GEX_CONTEXT_202607 | positive | AR | False | UNATTRIBUTED |  |
| GEX | E004 | GEX_CONTEXT_202607 | positive | ST | False | UNATTRIBUTED |  |
| GEX | E005 | GEX_CONTEXT_202607 | rejected | Spring | False | NO_PREDICTION |  |
| GEX | E006 | GEX_TR_202610 | rejected | AR | False | NO_PREDICTION |  |
| GEX | E007 | GEX_CONTEXT_202607 | rejected | SOS | False | NO_PREDICTION |  |
| GEX | E008 | GEX_TR_202610 | rejected | LPS | False | NO_PREDICTION |  |
| HDG | E001 | HDG_TR_202607 | uncertain |  | False | EXCLUDED_UNCERTAIN |  |
| HDG | E002 | HDG_TR_202607 | positive | AR | False | NO_PREDICTION |  |
| HDG | E003 | HDG_TR_202607 | positive | ST | False | NO_PREDICTION |  |
| HDG | E004 | HDG_TR_202607 | positive | SOS | True | MATCHED_TARGET | 2026-09-04 |
| HDG | E005 | HDG_TR_202607 | rejected | SOW | False | UNATTRIBUTED |  |
| HDG | E006 | HDG_TR_202610 | uncertain |  | False | EXCLUDED_UNCERTAIN |  |
| HDG | E007 | HDG_TR_202610 | positive | AR | False | NO_PREDICTION |  |
| HDG | E008 | HDG_TR_202610 | uncertain |  | False | EXCLUDED_UNCERTAIN |  |
| HDG | E009 | HDG_TR_202607 | rejected | Spring | False | UNMAPPED |  |
| HDG | E010 | HDG_TR_202610 | uncertain |  | False | EXCLUDED_UNCERTAIN |  |
| LPB | E001 | LPB_TR_202606 | positive | BC | False | NO_PREDICTION |  |
| LPB | E002 | LPB_TR_202606 | positive | AR | False | NO_PREDICTION |  |
| LPB | E003 | LPB_TR_202606 | positive | ST | False | NO_PREDICTION |  |
| LPB | E004 | LPB_TR_202606 | positive | SOW | False | UNATTRIBUTED |  |
| LPB | E005 | LPB_TR_202606 | positive | LPSY | False | NO_PREDICTION |  |
| LPB | E006 | LPB_MARKDOWN_202609 | rejected | SC | False | TARGET_UNDEFINED |  |
| LPB | E007 | LPB_MARKDOWN_202609 | rejected | SOS | False | TARGET_UNDEFINED |  |
| LPB | E008 | LPB_MARKDOWN_202609 | rejected | SC | False | TARGET_UNDEFINED |  |
| MSB | E001 | MSB_TR_202607 | positive | Spring | False | UNMAPPED |  |
| MSB | E002 | MSB_TR_202607 | uncertain |  | False | EXCLUDED_UNCERTAIN |  |
| MSB | E003 | MSB_TR_202607 | positive | SOS | True | MATCHED_TARGET | 2026-09-23 |
| MSB | E004 | MSB_TR_202607 | positive | LPS | False | UNMAPPED |  |
| MSB | E005 | MSB_TR_202607 | rejected | SC | False | NO_PREDICTION |  |
| MSB | E006 | MSB_TR_202607 | rejected | SOW | False | NO_PREDICTION |  |
| MSB | E007 | MSB_TR_202607 | rejected | LPS | False | NO_PREDICTION |  |
| NLG | E001 | NLG_TR_202609_SHORT | uncertain |  | False | EXCLUDED_UNCERTAIN |  |
| NLG | E002 | NLG_TR_202609_SHORT | positive | AR | False | NO_PREDICTION |  |
| NLG | E003 | NLG_TR_202609_SHORT | positive | Spring | True | MATCHED_TARGET | 2026-10-15 |
| NLG | E004 | NLG_TR_202609_SHORT | uncertain |  | False | EXCLUDED_UNCERTAIN |  |
| NLG | E005 | NLG_TR_202609_SHORT | positive | SOS | True | MATCHED_TARGET | 2026-10-21 |
| NLG | E006 | NLG_TR_202609_SHORT | uncertain |  | False | EXCLUDED_UNCERTAIN |  |
| NLG | E007 | NLG_TR_202609_SHORT | rejected | Test | False | NO_PREDICTION |  |
| NLG | E008 | NLG_CONTEXT_202607 | rejected | SOS | False | OTHER_STRUCTURE |  |
| NLG | E009 | NLG_CONTEXT_202607 | positive | SC | False | UNATTRIBUTED |  |
| NLG | E010 | NLG_CONTEXT_202607 | uncertain |  | False | EXCLUDED_UNCERTAIN |  |
| NLG | E011 | NLG_CONTEXT_202607 | positive | ST | False | UNATTRIBUTED |  |
| NVL | E001 | NVL_TR_202609 | uncertain |  | False | EXCLUDED_UNCERTAIN |  |
| NVL | E002 | NVL_TR_202609 | positive | SC | False | UNATTRIBUTED |  |
| NVL | E003 | NVL_TR_202609 | positive | AR | False | UNATTRIBUTED |  |
| NVL | E004 | NVL_TR_202609 | positive | ST | False | NO_PREDICTION |  |
| NVL | E005 | NVL_TR_202609 | rejected | SOS | False | NO_PREDICTION |  |
| NVL | E006 | NVL_TR_202609 | rejected | Spring | False | NO_PREDICTION |  |
| NVL | E007 | NVL_TR_202609 | rejected | LPS | False | NO_PREDICTION |  |
| SZC | E001 | SZC_TR_202607 | positive | PS | False | UNATTRIBUTED |  |
| SZC | E002 | SZC_TR_202607 | positive | SC | False | UNATTRIBUTED |  |
| SZC | E003 | SZC_TR_202607 | positive | AR | False | UNATTRIBUTED |  |
| SZC | E004 | SZC_TR_202607 | positive | ST | False | NO_PREDICTION |  |
| SZC | E005 | SZC_TR_202607 | positive | SOS | False | UNMAPPED |  |
| SZC | E006 | SZC_TR_202607 | uncertain |  | False | EXCLUDED_UNCERTAIN |  |
| SZC | E007 | SZC_TR_202607 | rejected | Spring | False | NO_PREDICTION |  |
| SZC | E008 | SZC_TR_202607 | rejected | SC | False | NO_PREDICTION |  |

## Interpretation and provenance

Validation run: https://github.com/trungvdn/investment-ai/actions/runs/41456103406 . Baseline analyzer: `4e27da49590b137fed4131bc34609a154706ecf1`; both versions use the identical scope-v1 evaluator and annotation files. Production analyzer behavior and reviewed labels are unchanged in this step. Equal before/after scoped scores are expected; this is an evaluation correction, not a claimed detection improvement.

Strict scoped recall is 6/36 (16.7%), versus the unchanged legacy date-only 21/36. These are different metrics. Ten of fourteen geometrically defined reviewed ranges have at least one candidate mapping during replay. Mapping coverage alone does not establish event correctness. There are 14 mapped, 103 unattributed, 55 unmapped, 7 low-confidence and 0 ambiguous unique predictions in the audit window. Unreviewed predictions cannot be treated as global false positives.

No rejected label is emitted on a confidently matched target under scope-v1. This is not zero-error proof: only 19 of 22 rejection targets have range geometry; the three LPB_MARKDOWN_202609 labels are TARGET_UNDEFINED, and legacy rejected SOW predictions can remain UNATTRIBUTED. Seventeen uncertain labels remain excluded. The two nested-range contradictions previously excluded by the legacy scorer are now evaluated against their own structure.

Important cases:

- NLG small-range Spring and SOS match NLG_TR_202609_SHORT. Its same-date SOS does not count against rejected SOS on NLG_CONTEXT_202607 (OTHER_STRUCTURE).
- MSB Spring and LPS match dates but their detected ranges remain UNMAPPED; only the SOS matches MSB_TR_202607. The previous date-only hits therefore did not prove correct attribution.
- CTD 17/09 Spring on 57.3–64.3 is LOW_CONFIDENCE relative to reviewed geometry, not automatically labelled correct or false on the wide range.
- HDG 09/10 Spring on 14.3–16.5 is UNMAPPED rather than a Spring of the rejected 15.2-support range. The rejected legacy SOW remains visible as unattributed evidence.

The fixed matching tolerances are experimental: <=5% of full reviewed range width beyond each reviewed zone, >=50% formation overlap, >=0.05 normalized cost separation. Proximity at 5–15% is held as LOW_CONFIDENCE. Automatic IDs are never equated by name, and event-label agreement does not influence geometry matching. Later reviewed geometry is used only for evaluation, never fed to Analyze; this remains a retrospective attribution audit.

Next priorities: assign PS/SC/AR/ST and remaining legacy events to defensible structures; repair MSB Spring/LPS range identity; define or explicitly exclude LPB trend targets before calculating a fully evaluable rejected-label rate. Keep unresolved mappings visible rather than inflating scores through nearest-range fallback.
