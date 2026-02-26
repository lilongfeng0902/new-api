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

import React, { useState } from 'react';
import { Card, Tabs, TabPane } from '@douyinfe/semi-ui';
import { VChart } from '@visactor/react-vchart';

const PublicChartsPanel = ({
  activeChartTab,
  setActiveChartTab,
  spec_line,
  spec_model_line,
  spec_pie,
  spec_rank_bar,
  spec_platform_trend,
  spec_top_users_bar,
  CARD_PROPS,
  hasApiInfoPanel,
  t,
}) => {
  const [activeTab, setActiveTab] = useState('1');

  const handleTabChange = (key) => {
    setActiveTab(key);
    setActiveChartTab && setActiveChartTab(key);
  };

  return (
    <div className={hasApiInfoPanel ? 'lg:col-span-3' : 'col-span-1'}>
      <Card {...CARD_PROPS} className='h-full'>
        <Tabs activeKey={activeTab} onChange={handleTabChange} type='line'>
          <TabPane tab={t('消耗趋势')} itemKey='1'>
            <div className='h-80'>
              <VChart spec={spec_line} />
            </div>
          </TabPane>
          <TabPane tab={t('模型趋势')} itemKey='2'>
            <div className='h-80'>
              <VChart spec={spec_model_line} />
            </div>
          </TabPane>
          <TabPane tab={t('模型占比')} itemKey='3'>
            <div className='h-80'>
              <VChart spec={spec_pie} />
            </div>
          </TabPane>
          <TabPane tab={t('模型排名')} itemKey='4'>
            <div className='h-80'>
              <VChart spec={spec_rank_bar} />
            </div>
          </TabPane>
          <TabPane tab={t('平台趋势')} itemKey='5'>
            <div className='h-80'>
              <VChart spec={spec_platform_trend} />
            </div>
          </TabPane>
          <TabPane tab={t('用户Top10')} itemKey='6'>
            <div className='h-80'>
              <VChart spec={spec_top_users_bar} />
            </div>
          </TabPane>
        </Tabs>
      </Card>
    </div>
  );
};

export default PublicChartsPanel;
