# Structure-scoped Wyckoff benchmark

scope-v1: boundary distance <=0.05 reviewed width each (0.05-0.15 held as LOW_CONFIDENCE); formation overlap >=0.5 shorter interval; cost separation >=0.05; no fallback for missing IDs

Evaluation only: annotations never enter Analyze. Automatic IDs are mapped by frozen boundaries and formation time, not by event-label agreement.
Unattributed, unmapped, low-confidence and ambiguous predictions are not counted as correct or rejected-target emissions. All reviewed positive labels remain in the denominator.
Reviewed candidates are hypotheses; sparse labels cannot establish global precision or investment accuracy.

| Metric | Baseline | Candidate |
|---|---:|---:|
| Positive | 36 | 36 |
| Hits | 6 | 8 |
| Rejected | 22 | 22 |
| RejectedEmitted | 0 | 0 |
| Uncertain | 17 | 17 |
| MappedPredictions | 14 | 19 |
| UnattributedPredictions | 103 | 97 |
| UnmappedPredictions | 55 | 72 |
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
| MSB | MSB_TR_202607 | True | 3/3 | 0/3 | 1 |
| NLG | NLG_CONTEXT_202607 | True | 0/2 | 0/1 | 1 |
| NLG | NLG_TR_202609_SHORT | True | 2/3 | 0/1 | 3 |
| NVL | NVL_TR_202609 | True | 0/3 | 0/3 | 1 |
| SZC | SZC_TR_202607 | False | 0/5 | 0/2 | 1 |

| Stock | Event | Reviewed structure | Kind | Label | Scoped hit | Disposition | First seen |
|---|---|---|---|---|---|---|---|
| BCM | E001 | BCM_CONTEXT_202607 | uncertain |  | False | EXCLUDED_UNCERTAIN |  |
| BCM | E002 | BCM_CONTEXT_202607 | positive | SC | False | UNATTRIBUTED |  |
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
| MSB | E001 | MSB_TR_202607 | positive | Spring | True | MATCHED_TARGET | 2026-08-27 |
| MSB | E002 | MSB_TR_202607 | uncertain |  | False | EXCLUDED_UNCERTAIN |  |
| MSB | E003 | MSB_TR_202607 | positive | SOS | True | MATCHED_TARGET | 2026-09-23 |
| MSB | E004 | MSB_TR_202607 | positive | LPS | True | MATCHED_TARGET | 2026-10-05 |
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

## Primary-range chart fix

Baseline: `21f31ac276c3de7b8074009538a01045fd8d8d14`. Candidate validation: https://github.com/trungvdn/investment-ai/actions/runs/41459184684 . Raw date-only recall changes 21/36 -> 22/36; strict scope-v1 recall changes 6/36 -> 8/36. Raw rejected-date overlaps remain 3/20; confidently mapped rejected targets remain zero, with the existing missing-attribution/undefined-target limitations. These sparse metrics do not establish overall precision.

The two additional scoped hits are MSB Spring and LPS now attached to the same reviewed primary range as SOS. SC selection adds BCM 27/07 while preserving NVL's 30/09 high-volume absorption climax. Chart/report curation is separate from raw candidate evaluation; omitted candidates are not silently removed from benchmark inventories.

Expected primary chart timeline from reviewed OHLCV through 30/10/2026:

| Symbol | Visible events |
|---|---|
| MSB | Spring 26/08, SOS 23/09, LPS 05/10; no promoted TEST, one SOS/LPS |
| BCM | PS 23/07, SC 27/07, AR 13/08, ST 07/10 and 14/10; no smaller-range Spring |

Raw multi-range candidates remain available for diagnostics but do not populate the primary chart. Legacy stopping events such as BCM PS/SC/AR/ST still lack automatic structure IDs; their presentation is gated by primary price boundaries and sequence checks, not falsely credited as scoped benchmark hits. Reviewed annotations remain unchanged. Fixture tests do not guarantee all future live labels.
