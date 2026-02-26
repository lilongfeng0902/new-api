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

import React from 'react';
import { Card, Skeleton } from '@douyinfe/semi-ui';
import { Database, Activity, Zap, BarChart3, Boxes, Building2, Key, TrendingUp } from 'lucide-react';
import { renderQuota } from '../../helpers';

const PublicStatsCards = ({ quotaData, publicStats, loading, t, CARD_PROPS }) => {
  // 从publicStats获取统计数据，如果不存在则从quotaData计算（向后兼容）
  const statsData = React.useMemo(() => {
    const totalRequests = publicStats?.total_req_count ?? (quotaData?.reduce((sum, item) => sum + (item.count || 0), 0) || 0);
    const totalQuota = publicStats?.total_quota ?? (quotaData?.reduce((sum, item) => sum + (item.quota || 0), 0) || 0);
    const totalTokens = publicStats?.total_token_usage ?? (quotaData?.reduce((sum, item) => sum + (item.token_used || 0), 0) || 0);
    const dataCount = publicStats?.total_data_count ?? (quotaData?.length || 0);

    return [
      {
        title: t('总请求次数'),
        value: totalRequests.toLocaleString(),
        icon: <Activity />,
        avatarColor: 'green',
      },
      {
        title: t('总消耗额度'),
        value: renderQuota(totalQuota),
        icon: <BarChart3 />,
        avatarColor: 'purple',
      },
      {
        title: t('总Token消耗'),
        value: totalTokens.toLocaleString(),
        icon: <Zap />,
        avatarColor: 'orange',
      },
      {
        title: t('数据记录数'),
        value: dataCount.toLocaleString(),
        icon: <Database />,
        avatarColor: 'blue',
      },
      {
        title: t('接入模型数'),
        value: publicStats?.enabled_models_count?.toLocaleString() || '0',
        icon: <Boxes />,
        avatarColor: 'cyan',
      },
      {
        title: t('接入渠道数'),
        value: publicStats?.enabled_channels_count?.toLocaleString() || '0',
        icon: <Building2 />,
        avatarColor: 'indigo',
      },
      {
        title: t('令牌总数'),
        value: publicStats?.active_tokens_count?.toLocaleString() || '0',
        icon: <Key />,
        avatarColor: 'pink',
      },
      {
        title: t('今日Token消耗'),
        value: publicStats?.today_token_usage?.toLocaleString() || '0',
        icon: <TrendingUp />,
        avatarColor: 'red',
      },
    ];
  }, [quotaData, publicStats, t]);

  return (
    <div className='mb-6'>
      <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4'>
        {statsData.map((item, index) => (
          <Card
            key={index}
            {...CARD_PROPS}
            className='relative overflow-hidden'
          >
            {loading ? (
              <Skeleton
                placeholder={
                  <div>
                    <Skeleton.Title style={{ width: 80 }} />
                    <Skeleton.Paragraph rows={1} style={{ width: 60 }} />
                  </div>
                }
                loading={loading}
              />
            ) : (
              <div className='flex items-center space-x-3'>
                <div
                  className={`p-2 rounded-lg bg-${item.avatarColor}-100 dark:bg-${item.avatarColor}-900/20`}
                >
                  <div
                    className={`text-${item.avatarColor}-600 dark:text-${item.avatarColor}-400`}
                  >
                    {item.icon}
                  </div>
                </div>
                <div className='flex-1 min-w-0'>
                  <p className='text-sm font-medium text-gray-600 dark:text-gray-400'>
                    {item.title}
                  </p>
                  <p className='text-xl font-bold text-gray-900 dark:text-gray-100 truncate'>
                    {item.value}
                  </p>
                </div>
              </div>
            )}
          </Card>
        ))}
      </div>
    </div>
  );
};

export default PublicStatsCards;
