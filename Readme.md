# OpenTelemetry
### 简介
OpenTrace(链路追踪标准) + OpenCensus(指标\监控标准)= OpenTelemetry

#### metrics(指标/度量)
1. counter(一个随时间累加的值，不会变)
2. measure(随时间聚合的值，直方图)
3. observer(捕获特定时间点的当前值，实时)
#### baggage
通过context传输数据