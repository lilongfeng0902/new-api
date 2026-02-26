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

import { useState, useEffect, useRef, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { API, showError, timestamp2string } from '../../helpers';
import { getRelativeTime } from '../../helpers';
import { useMinimumLoadingTime } from '../common/useMinimumLoadingTime';

export const usePublicDashboardData = (statusState) => {
  const { t } = useTranslation();
  const initialized = useRef(false);

  // ========== 基础状态 ==========
  const [loading, setLoading] = useState(false);
  const showLoading = useMinimumLoadingTime(loading);

  // ========== 输入状态 ==========
  const [inputs, setInputs] = useState({
    start_timestamp: timestamp2string(
      (new Date().getTime() - 7 * 24 * 3600 * 1000) / 1000,
    ),
    end_timestamp: timestamp2string(new Date().getTime() / 1000),
  });

  const [dataExportDefaultTime, setDataExportDefaultTime] = useState('hour');

  // ========== 数据状态 ==========
  const [quotaData, setQuotaData] = useState([]);
  const [publicStats, setPublicStats] = useState({
    enabled_models_count: 0,
    enabled_channels_count: 0,
    active_tokens_count: 0,
    today_token_usage: 0,
    total_req_count: 0,
    total_quota: 0,
    total_token_usage: 0,
    total_data_count: 0,
  });
  const [topUsers, setTopUsers] = useState([]);

  // ========== 图表状态 ==========
  const [activeChartTab, setActiveChartTab] = useState('1');

  // ========== Uptime 数据 ==========
  const [uptimeData, setUptimeData] = useState([]);
  const [uptimeLoading, setUptimeLoading] = useState(false);
  const [activeUptimeTab, setActiveUptimeTab] = useState('');

  // ========== Panel enable flags ==========
  const apiInfoEnabled = statusState?.status?.api_info_enabled ?? true;
  const announcementsEnabled =
    statusState?.status?.announcements_enabled ?? true;
  const faqEnabled = statusState?.status?.faq_enabled ?? true;
  const uptimeEnabled = statusState?.status?.uptime_kuma_enabled ?? true;

  const hasApiInfoPanel = apiInfoEnabled;
  const hasInfoPanels = announcementsEnabled || faqEnabled || uptimeEnabled;

  // ========== API 调用函数 ==========
  const loadPublicData = useCallback(async () => {
    setLoading(true);
    try {
      const { start_timestamp, end_timestamp } = inputs;
      let localStartTimestamp = Date.parse(start_timestamp) / 1000;
      let localEndTimestamp = Date.parse(end_timestamp) / 1000;

      const url = `/api/data/public/?start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}`;

      const res = await API.get(url);
      const { success, message, data } = res.data;
      if (success) {
        setQuotaData(data || []);
        return data;
      } else {
        showError(message);
        return null;
      }
    } finally {
      setLoading(false);
    }
  }, [inputs]);

  const loadPublicStats = useCallback(async () => {
    try {
      const { start_timestamp, end_timestamp } = inputs;
      let localStartTimestamp = Date.parse(start_timestamp) / 1000;
      let localEndTimestamp = Date.parse(end_timestamp) / 1000;

      const url = `/api/data/public/stats?start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}`;

      const res = await API.get(url);
      const { success, message, data } = res.data;
      if (success) {
        setPublicStats(data.stats || {});
        setTopUsers(data.top_users || []);
        return data;
      } else {
        showError(message);
        return null;
      }
    } catch (err) {
      console.error('Failed to load public stats:', err);
      return null;
    }
  }, [inputs]);

  const loadUptimeData = useCallback(async () => {
    setUptimeLoading(true);
    try {
      const res = await API.get('/api/uptime/status');
      const { success, message, data } = res.data;
      if (success) {
        setUptimeData(data || []);
        if (data && data.length > 0 && !activeUptimeTab) {
          setActiveUptimeTab(data[0].categoryName);
        }
      } else {
        showError(message);
      }
    } catch (err) {
      console.error(err);
    } finally {
      setUptimeLoading(false);
    }
  }, [activeUptimeTab]);

  const refresh = useCallback(async () => {
    const data = await loadPublicData();
    await loadPublicStats();
    await loadUptimeData();
    return data;
  }, [loadPublicData, loadPublicStats, loadUptimeData]);

  // ========== Effects ==========
  useEffect(() => {
    if (!initialized.current) {
      loadPublicData();
      initialized.current = true;
    }
  }, [loadPublicData]);

  // ========== 工具函数 ==========
  const handleCopyUrl = useCallback((url) => {
    navigator.clipboard.writeText(url);
  }, []);

  const handleSpeedTest = useCallback(() => {
    // TODO: 实现速度测试功能
    console.log('Speed test not implemented yet');
  }, []);

  const renderMonitorList = useCallback(
    (monitors, getUptimeStatusColor, getUptimeStatusText, t) => {
      return monitors.map((monitor, index) => (
        <div
          key={index}
          className='flex items-center justify-between p-2 rounded-lg bg-gray-50 dark:bg-gray-800'
        >
          <div className='flex items-center space-x-2'>
            <div
              className='w-3 h-3 rounded-full'
              style={{ backgroundColor: getUptimeStatusColor(monitor.status) }}
            />
            <span className='text-sm font-medium'>{monitor.name}</span>
          </div>
          <span className='text-sm text-gray-600 dark:text-gray-400'>
            {getUptimeStatusText(monitor.status, t)}
          </span>
        </div>
      ));
    },
    [],
  );

  const getUptimeStatusColor = useCallback((status) => {
    const statusColors = {
      0: '#10b981', // online
      1: '#f59e0b', // maintenance
      2: '#ef4444', // offline
    };
    return statusColors[status] || '#6b7280';
  }, []);

  const getUptimeStatusText = useCallback(
    (status, t) => {
      const statusTexts = {
        0: t('正常'),
        1: t('维护中'),
        2: t('离线'),
      };
      return statusTexts[status] || t('未知');
    },
    [t],
  );

  return {
    // 基础状态
    loading: showLoading,

    // 输入状态
    inputs,
    dataExportDefaultTime,

    // 数据状态
    quotaData,
    publicStats,
    topUsers,

    // 图表状态
    activeChartTab,
    setActiveChartTab,

    // Uptime 数据
    uptimeData,
    uptimeLoading,
    activeUptimeTab,
    setActiveUptimeTab,

    // 计算值
    hasApiInfoPanel,
    hasInfoPanels,
    apiInfoEnabled,
    announcementsEnabled,
    faqEnabled,
    uptimeEnabled,

    // 函数
    loadPublicData,
    loadPublicStats,
    loadUptimeData,
    refresh,

    // 工具函数
    getRelativeTime,
    handleCopyUrl,
    handleSpeedTest,
    renderMonitorList,
    getUptimeStatusColor,
    getUptimeStatusText,

    // 翻译
    t,
  };
};
