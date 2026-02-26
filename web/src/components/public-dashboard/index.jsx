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

import React, { useContext, useEffect } from 'react';
import { StatusContext } from '../../context/Status';

import PublicDashboardHeader from './PublicDashboardHeader';
import PublicStatsCards from './PublicStatsCards';
import PublicChartsPanel from './PublicChartsPanel';
import ApiInfoPanel from '../dashboard/ApiInfoPanel';
import AnnouncementsPanel from '../dashboard/AnnouncementsPanel';
import FaqPanel from '../dashboard/FaqPanel';
import UptimePanel from '../dashboard/UptimePanel';

import { usePublicDashboardData } from '../../hooks/public-dashboard/usePublicDashboardData';
import { usePublicDashboardCharts } from '../../hooks/public-dashboard/usePublicDashboardCharts';

import {
  CARD_PROPS,
  FLEX_CENTER_GAP2,
  ILLUSTRATION_SIZE,
  ANNOUNCEMENT_LEGEND_DATA,
  UPTIME_STATUS_MAP,
} from '../../constants/dashboard.constants';

const PublicDashboard = () => {
  const [statusState, statusDispatch] = useContext(StatusContext);

  // ========== 主要数据管理 ==========
  const publicDashboardData = usePublicDashboardData(statusState);

  // ========== 图表管理 ==========
  const publicDashboardCharts = usePublicDashboardCharts(publicDashboardData.t);

  // ========== 数据处理 ==========
  const initChart = async () => {
    await publicDashboardData.loadPublicData().then((data) => {
      if (data && data.length > 0) {
        publicDashboardCharts.updateChartData(data);
      }
    });
    await publicDashboardData.loadPublicStats().then((data) => {
      if (data && data.top_users && data.top_users.length > 0) {
        publicDashboardCharts.updateTopUsersChart(data.top_users);
      }
    });
    await publicDashboardData.loadUptimeData();
  };

  const handleRefresh = async () => {
    const data = await publicDashboardData.refresh();
    if (data && data.length > 0) {
      publicDashboardCharts.updateChartData(data);
    }
    const statsData = await publicDashboardData.loadPublicStats();
    if (statsData && statsData.top_users && statsData.top_users.length > 0) {
      publicDashboardCharts.updateTopUsersChart(statsData.top_users);
    }
  };

  // ========== 数据准备 ==========
  const apiInfoData = statusState?.status?.api_info || [];
  const announcementData = (statusState?.status?.announcements || []).map(
    (item) => {
      const pubDate = item?.publishDate ? new Date(item.publishDate) : null;
      const absoluteTime =
        pubDate && !isNaN(pubDate.getTime())
          ? `${pubDate.getFullYear()}-${String(pubDate.getMonth() + 1).padStart(2, '0')}-${String(pubDate.getDate()).padStart(2, '0')} ${String(pubDate.getHours()).padStart(2, '0')}:${String(pubDate.getMinutes()).padStart(2, '0')}`
          : item?.publishDate || '';
      const relativeTime = publicDashboardData.getRelativeTime(
        item.publishDate,
      );
      return {
        ...item,
        time: absoluteTime,
        relative: relativeTime,
      };
    },
  );
  const faqData = statusState?.status?.faq || [];

  const uptimeLegendData = Object.entries(UPTIME_STATUS_MAP).map(
    ([status, info]) => ({
      status: Number(status),
      color: info.color,
      label: publicDashboardData.t(info.label),
    }),
  );

  // ========== Effects ==========
  useEffect(() => {
    initChart();
  }, []);

  return (
    <div className='h-full'>
      <PublicDashboardHeader
        refresh={handleRefresh}
        loading={publicDashboardData.loading}
        t={publicDashboardData.t}
      />

      <PublicStatsCards
        quotaData={publicDashboardData.quotaData}
        publicStats={publicDashboardData.publicStats}
        loading={publicDashboardData.loading}
        t={publicDashboardData.t}
        CARD_PROPS={CARD_PROPS}
      />

      {/* API信息和图表面板 */}
      <div className='mb-4'>
        <div
          className={`grid grid-cols-1 gap-4 ${publicDashboardData.hasApiInfoPanel ? 'lg:grid-cols-4' : ''}`}
        >
          <PublicChartsPanel
            activeChartTab={publicDashboardData.activeChartTab}
            setActiveChartTab={publicDashboardData.setActiveChartTab}
            spec_line={publicDashboardCharts.spec_line}
            spec_model_line={publicDashboardCharts.spec_model_line}
            spec_pie={publicDashboardCharts.spec_pie}
            spec_rank_bar={publicDashboardCharts.spec_rank_bar}
            spec_platform_trend={publicDashboardCharts.spec_platform_trend}
            spec_top_users_bar={publicDashboardCharts.spec_top_users_bar}
            CARD_PROPS={CARD_PROPS}
            hasApiInfoPanel={publicDashboardData.hasApiInfoPanel}
            t={publicDashboardData.t}
          />

          {publicDashboardData.hasApiInfoPanel && (
            <ApiInfoPanel
              apiInfoData={apiInfoData}
              handleCopyUrl={(url) => publicDashboardData.handleCopyUrl(url)}
              handleSpeedTest={publicDashboardData.handleSpeedTest}
              CARD_PROPS={CARD_PROPS}
              FLEX_CENTER_GAP2={FLEX_CENTER_GAP2}
              ILLUSTRATION_SIZE={ILLUSTRATION_SIZE}
              t={publicDashboardData.t}
            />
          )}
        </div>
      </div>

      {/* 系统公告和常见问答卡片 */}
      {publicDashboardData.hasInfoPanels && (
        <div className='mb-4'>
          <div className='grid grid-cols-1 lg:grid-cols-4 gap-4'>
            {/* 公告卡片 */}
            {publicDashboardData.announcementsEnabled && (
              <AnnouncementsPanel
                announcementData={announcementData}
                announcementLegendData={ANNOUNCEMENT_LEGEND_DATA.map(
                  (item) => ({
                    ...item,
                    label: publicDashboardData.t(item.label),
                  }),
                )}
                CARD_PROPS={CARD_PROPS}
                ILLUSTRATION_SIZE={ILLUSTRATION_SIZE}
                t={publicDashboardData.t}
              />
            )}

            {/* 常见问答卡片 */}
            {publicDashboardData.faqEnabled && (
              <FaqPanel
                faqData={faqData}
                CARD_PROPS={CARD_PROPS}
                FLEX_CENTER_GAP2={FLEX_CENTER_GAP2}
                ILLUSTRATION_SIZE={ILLUSTRATION_SIZE}
                t={publicDashboardData.t}
              />
            )}

            {/* 服务可用性卡片 */}
            {publicDashboardData.uptimeEnabled && (
              <UptimePanel
                uptimeData={publicDashboardData.uptimeData}
                uptimeLoading={publicDashboardData.uptimeLoading}
                activeUptimeTab={publicDashboardData.activeUptimeTab}
                setActiveUptimeTab={publicDashboardData.setActiveUptimeTab}
                loadUptimeData={publicDashboardData.loadUptimeData}
                uptimeLegendData={uptimeLegendData}
                renderMonitorList={(monitors) =>
                  publicDashboardData.renderMonitorList(
                    monitors,
                    publicDashboardData.getUptimeStatusColor,
                    publicDashboardData.getUptimeStatusText,
                    publicDashboardData.t,
                  )
                }
                CARD_PROPS={CARD_PROPS}
                ILLUSTRATION_SIZE={ILLUSTRATION_SIZE}
                t={publicDashboardData.t}
              />
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default PublicDashboard;
