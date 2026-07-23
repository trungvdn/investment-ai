# Market Intelligence

## 1. Purpose

Market Intelligence is responsible for understanding the overall state of the financial market by collecting, analyzing, and synthesizing market-wide information into actionable investment insights.

Rather than focusing on individual securities, this capability provides a macro view of market conditions, helping investors answer questions such as:

* What is happening in the market today?
* Is the market healthy?
* What are the dominant trends?
* Where is money flowing?
* What are the major risks?

This capability serves as one of the foundational inputs for portfolio construction and investment decision-making.

---

# 2. Business Value

Market Intelligence enables investors to:

* Understand the current market environment.
* Identify market trends early.
* Detect abnormal market behavior.
* Monitor liquidity and capital flows.
* Reduce research time.
* Improve investment decision quality.

---

# 3. Responsibilities

This capability is responsible for:

* Monitoring overall market performance.
* Tracking market breadth.
* Measuring market momentum.
* Measuring liquidity.
* Monitoring sector rotation.
* Detecting institutional money flow.
* Evaluating market sentiment.
* Detecting unusual market events.
* Producing explainable market insights.

This capability is **not responsible** for making investment decisions.

---

# 4. Inputs

## Market Data

* Market Index
* Index Constituents
* Price
* Volume
* Order Book
* Historical Prices
* Trading Value

## Technical Indicators

* Moving Averages
* RSI
* MACD
* Bollinger Bands
* ATR
* ADX

## Market Breadth

* Advance / Decline
* New High / New Low
* Up Volume / Down Volume
* Market Participation

## Sector Data

* Sector Performance
* Sector Rotation
* Relative Strength

## Capital Flow

* Foreign Investors
* Institutional Investors
* ETF Flows
* Proprietary Trading

## Alternative Data

* News
* Economic Calendar
* Market Events

---

# 5. Outputs

This capability produces structured market intelligence including:

* Market Trend
* Market Health
* Market Breadth
* Liquidity Analysis
* Sector Rotation
* Money Flow
* Risk Assessment
* Market Sentiment
* Confidence Score
* Explainable Reasoning

Outputs should be consumable by:

* Portfolio Intelligence
* Risk Intelligence
* Report Generation
* Alerting
* AI Agents
* Human Users

---

# 6. Business Rules

Examples:

### Rule 1

If most sectors are rising while liquidity increases,
market strength should be considered improving.

---

### Rule 2

If the index rises but market breadth deteriorates,
flag a potential divergence.

---

### Rule 3

Large institutional buying should increase confidence,
but should not independently generate buy recommendations.

---

### Rule 4

Every insight must include supporting evidence.

---

### Rule 5

Every conclusion must include an explanation.

---

# 7. Dependencies

Internal Capabilities

* News Intelligence
* Macro Intelligence
* Sector Intelligence
* Knowledge Management

External Systems

* Market Data Provider
* News Provider
* Economic Calendar
* Financial Database

---

# 8. Non-Goals

This capability does not:

* Execute trades.
* Recommend exact buy/sell orders.
* Manage portfolios.
* Predict future prices with certainty.

---

# 9. Future Enhancements

Potential future improvements include:

* AI-powered market regime detection.
* Market anomaly detection.
* Real-time event correlation.
* Cross-market analysis.
* Global market monitoring.
* Market forecasting.
* Reflection-based continuous learning.

---

# 10. Success Metrics

Examples:

* Market trend classification accuracy.
* Market regime detection accuracy.
* Insight generation latency.
* Explainability quality.
* User satisfaction.
* Reduction in research time.

---

# 11. Related Workflows

* Daily Market Analysis
* Weekly Market Review
* Market Alert
* Investment Research
* Portfolio Review

---

# 12. Related ADRs

* Workflow First
* Capability Driven Design
* Event Driven Architecture
* Knowledge Architecture
* Memory Strategy

---

# 13. Open Questions

* How should market sentiment be measured?
* Should market regimes be rule-based or AI-based?
* How should confidence scores be calculated?
* How should conflicting signals be resolved?
* How much historical data should influence current analysis?
