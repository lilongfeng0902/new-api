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
import { Button, Spin } from '@douyinfe/semi-ui';
import { IconRefresh } from '@douyinfe/semi-icons';

const PublicDashboardHeader = ({ refresh, loading, t }) => {
  return (
    <div className='mb-4 flex items-center justify-between'>
      <div>
        <h1 className='text-2xl font-bold text-gray-900 dark:text-gray-100'>
          {t('系统数据看板')}
        </h1>
        <p className='text-gray-600 dark:text-gray-400 mt-1'>
          {t('查看系统整体运行数据和使用统计')}
        </p>
      </div>
      <div className='flex items-center space-x-2'>
        <Button
          icon={<IconRefresh />}
          onClick={refresh}
          loading={loading}
          theme='light'
          type='tertiary'
        >
          {t('刷新')}
        </Button>
      </div>
    </div>
  );
};

export default PublicDashboardHeader;
