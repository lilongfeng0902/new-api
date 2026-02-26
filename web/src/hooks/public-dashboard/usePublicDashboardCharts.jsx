/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import { useState, useCallback } from 'react';
import {
  renderNumber,
  modelToColor,
} from '../../helpers';
import {
  updateChartSpec,
} from '../../helpers/dashboard';

export const usePublicDashboardCharts = (t) => {
  // ========== 数据状态 ==========
  const [trendData, setTrendData] = useState({
    balance: [],
    usedQuota: [],
    requestCount: [],
    times: [],
    consumeQuota: [],
    tokens: [],
    rpm: [],
    tpm: [],
  });
  const [consumeQuota, setConsumeQuota] = useState(0);
  const [times, setTimes] = useState(0);
  const [consumeTokens, setConsumeTokens] = useState(0);
  const [pieData, setPieData] = useState([]);
  const [lineData, setLineData] = useState([]);
  const [modelColors, setModelColors] = useState({});
  const [platformTrendData, setPlatformTrendData] = useState([]);
  const [topUsersData, setTopUsersData] = useState([]);

  // ========== 图表规格状态 ==========
  const [spec_pie, setSpecPie] = useState({
    type: 'pie',
    data: [
      {
        id: 'id0',
        values: [{ type: 'null', value: '0' }],
      },
    ],
    outerRadius: 0.8,
    innerRadius: 0.5,
    padAngle: 0.6,
    valueField: 'value',
    categoryField: 'type',
    pie: {
      style: {
        cornerRadius: 10,
      },
      state: {
        hover: {
          outerRadius: 0.85,
          stroke: '#000',
          lineWidth: 1,
        },
        selected: {
          outerRadius: 0.85,
          stroke: '#000',
          lineWidth: 1,
        },
      },
    },
    title: {
      visible: true,
      text: t('模型调用次数占比'),
      subtext: `${t('总计')}：${renderNumber(0)}`,
    },
    legends: {
      visible: true,
      orient: 'left',
    },
    label: {
      visible: true,
    },
    tooltip: {
      mark: {
        content: [
          {
            key: (datum) => datum['type'],
            value: (datum) => renderNumber(datum['value']),
          },
        ],
      },
    },
  });

  const [spec_line, setSpecLine] = useState({
    type: 'line',
    data: [
      {
        id: 'line',
        values: [],
      },
    ],
    xField: 'time',
    yField: 'value',
    seriesField: 'type',
    point: {
      visible: false,
    },
    line: {
      style: {
        lineWidth: 2,
      },
    },
    axes: [
      {
        orient: 'bottom',
        type: 'band',
      },
      {
        orient: 'left',
        type: 'linear',
      },
    ],
    title: {
      visible: true,
      text: t('消耗趋势'),
    },
    legends: {
      visible: true,
      orient: 'top',
    },
    tooltip: {
      visible: true,
    },
  });

  const [spec_model_line, setSpecModelLine] = useState({
    type: 'line',
    data: [
      {
        id: 'lineData',
        values: [],
      },
    ],
    xField: 'Time',
    yField: 'Count',
    seriesField: 'Model',
    legends: {
      visible: true,
      selectMode: 'single',
    },
    title: {
      visible: true,
      text: t('模型消耗趋势'),
      subtext: '',
    },
    point: {
      visible: false,
    },
    line: {
      style: {
        lineWidth: 2,
      },
    },
    axes: [
      {
        orient: 'bottom',
        type: 'band',
      },
      {
        orient: 'left',
        type: 'linear',
      },
    ],
    tooltip: {
      visible: true,
    },
  });

  const [spec_rank_bar, setSpecRankBar] = useState({
    type: 'bar',
    data: [
      {
        id: 'rankData',
        values: [],
      },
    ],
    xField: 'Model',
    yField: 'Count',
    seriesField: 'Model',
    legends: {
      visible: true,
      selectMode: 'single',
    },
    title: {
      visible: true,
      text: t('模型调用次数排行'),
      subtext: '',
    },
    bar: {
      state: {
        hover: {
          stroke: '#000',
          lineWidth: 1,
        },
        selected: {
          fill: '#000',
        },
      },
    },
    axes: [
      {
        orient: 'bottom',
        type: 'band',
      },
      {
        orient: 'left',
        type: 'linear',
      },
    ],
    tooltip: {
      visible: true,
    },
  });

  const [spec_platform_trend, setSpecPlatformTrend] = useState({
    type: 'line',
    data: [{ id: 'platformTrend', values: [] }],
    xField: 'time',
    yField: 'quota',
    point: { visible: true, style: { size: 4 } },
    line: {
      style: {
        lineWidth: 3,
        stroke: '#8b5cf6',
      },
    },
    area: {
      visible: true,
      style: {
        fill: 'l(90) 0:#8b5cf6 1:#ffffff',
        fillOpacity: 0.2,
      },
    },
    title: {
      visible: true,
      text: t('平台消费趋势'),
      subtext: '',
    },
    axes: [
      {
        orient: 'bottom',
        type: 'band',
        label: { autoRotate: true },
      },
      {
        orient: 'left',
        type: 'linear',
        label: {
          formatMethod: (val) => renderNumber(val),
        },
      },
    ],
    legends: { visible: false },
    tooltip: {
      mark: {
        content: [
          {
            key: () => t('时间'),
            value: (datum) => datum['time'],
          },
          {
            key: () => t('消耗额度'),
            value: (datum) => renderNumber(datum['quota']),
          },
          {
            key: () => t('请求次数'),
            value: (datum) => datum['count']?.toLocaleString(),
          },
        ],
      },
    },
  });

  const [spec_top_users_bar, setSpecTopUsersBar] = useState({
    type: 'bar',
    data: [{ id: 'topUsersData', values: [] }],
    xField: 'username',
    yField: 'quota',
    seriesField: 'username',
    direction: 'horizontal',
    legends: { visible: false },
    title: {
      visible: true,
      text: t('用户消费Top10'),
      subtext: '',
    },
    bar: {
      style: { cornerRadius: 4 },
      state: {
        hover: { stroke: '#000', lineWidth: 1 },
      },
    },
    axes: [
      {
        orient: 'left',
        type: 'band',
        label: { autoLimit: true },
      },
      {
        orient: 'bottom',
        type: 'linear',
        label: {
          formatMethod: (val) => renderNumber(val),
        },
      },
    ],
    tooltip: {
      mark: {
        content: [
          {
            key: () => t('用户'),
            value: (datum) => datum['username'],
          },
          {
            key: () => t('消耗额度'),
            value: (datum) => renderNumber(datum['quota']),
          },
          {
            key: () => t('Token消耗'),
            value: (datum) => datum['token_used']?.toLocaleString(),
          },
          {
            key: () => t('请求次数'),
            value: (datum) => datum['request_count']?.toLocaleString(),
          },
        ],
      },
    },
  });

  // ========== 数据更新函数 ==========
  const updateChartData = useCallback(
    (data) => {
      if (!data || data.length === 0) {
        return;
      }

      // 确保 t 是一个函数
      if (typeof t !== 'function') {
        console.error('t is not a function:', t);
        return;
      }

      try {
        // 直接处理原始数据，避免复杂的 processRawData 调用
        if (!Array.isArray(data) || data.length === 0) {
          console.warn('No data to process');
          return;
        }

        // 计算基本统计数据
        let totalQuota = 0;
        let totalTimes = 0;
        let totalTokens = 0;
        const modelMap = new Map();

        data.forEach((item) => {
          totalQuota += item.quota || 0;
          totalTimes += item.count || 0;
          totalTokens += item.token_used || 0;

          // 统计每个模型的数据
          const modelName = item.model_name;
          if (!modelMap.has(modelName)) {
            modelMap.set(modelName, { count: 0, quota: 0 });
          }
          const modelData = modelMap.get(modelName);
          modelData.count += item.count || 0;
          modelData.quota += item.quota || 0;
        });

        setConsumeQuota(totalQuota);
        setTimes(totalTimes);
        setConsumeTokens(totalTokens);

        // 处理饼图数据
        const pieData = Array.from(modelMap.entries())
          .map(([model, data]) => ({
            type: model,
            value: data.count,
          }))
          .sort((a, b) => b.value - a.value);
        setPieData(pieData);

        // 处理折线图数据 - 按时间分组
        const timeMap = new Map();
        data.forEach((item) => {
          const timeKey = new Date(item.created_at * 1000).toLocaleDateString();
          if (!timeMap.has(timeKey)) {
            timeMap.set(timeKey, { quota: 0, count: 0 });
          }
          const timeData = timeMap.get(timeKey);
          timeData.quota += item.quota || 0;
          timeData.count += item.count || 0;
        });

        const lineData = Array.from(timeMap.entries()).map(([time, data]) => ({
          time,
          value: data.quota,
          type: '消耗额度',
        }));
        setLineData(lineData);

        // 处理平台总体趋势数据（按时间聚合所有模型）
        const platformTrend = Array.from(timeMap.entries())
          .map(([time, data]) => ({
            time,
            quota: data.quota,
            count: data.count,
          }))
          .sort((a, b) => new Date(a.time) - new Date(b.time));

        setPlatformTrendData(platformTrend);

        updateChartSpec(
          setSpecPlatformTrend,
          platformTrend,
          `${t('数据点')}：${platformTrend.length}`,
          {},
          'platformTrend'
        );

        // 设置模型颜色
        const modelColors = {};
        Array.from(modelMap.keys()).forEach((model) => {
          modelColors[model] = modelToColor(model);
        });
        setModelColors(modelColors);

        // 简化趋势数据（暂时使用空数据）
        setTrendData({
          balance: [],
          usedQuota: [],
          requestCount: [],
          times: [],
          consumeQuota: [],
          tokens: [],
          rpm: [],
          tpm: [],
        });

        // 更新图表规格
        updateChartSpec(
          setSpecPie,
          pieData,
          `${t('总计')}：${totalTimes}`,
          modelColors,
          'id0'
        );
        updateChartSpec(
          setSpecLine,
          lineData,
          `${t('总计')}：${totalQuota}`,
          modelColors,
          'line'
        );

        // 为模型趋势图准备正确格式的数据
        const modelLineData = Array.from(modelMap.entries()).map(([model, modelData]) => ({
          Time: '汇总',
          Count: modelData.count,
          Model: model,
        }));

        updateChartSpec(
          setSpecModelLine,
          modelLineData,
          `${t('数据点数')}：${data.length}`,
          modelColors,
          'lineData'
        );
        // 为排名图准备正确格式的数据
        const rankData = Array.from(modelMap.entries()).map(([model, modelData]) => ({
          Model: model,
          Count: modelData.count,
        }));

        updateChartSpec(
          setSpecRankBar,
          rankData,
          `${t('模型数')}：${rankData.length}`,
          modelColors,
          'rankData'
        );
      } catch (error) {
        console.error('Error updating chart data:', error);
      }
    },
    [t],
  );

  const updateTopUsersChart = useCallback((topUsersData) => {
    if (!topUsersData || topUsersData.length === 0) {
      return;
    }

    setTopUsersData(topUsersData);

    updateChartSpec(
      setSpecTopUsersBar,
      topUsersData,
      `${t('用户数')}：${topUsersData.length}`,
      {},
      'topUsersData'
    );
  }, [t]);

  return {
    // 图表规格
    spec_pie,
    spec_line,
    spec_model_line,
    spec_rank_bar,
    spec_platform_trend,
    spec_top_users_bar,

    // 数据状态
    trendData,
    consumeQuota,
    times,
    consumeTokens,
    pieData,
    lineData,
    modelColors,
    platformTrendData,
    topUsersData,

    // 函数
    updateChartData,
    updateTopUsersChart,
  };
};
