import { useState, useEffect } from 'react';
import { DataGrid, GridColDef, GridRowParams, GridSortModel, GridRenderCellParams } from '@mui/x-data-grid';
import { IconButton } from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';
import { useMutation, useQueryClient } from 'react-query';
import axios from 'axios';
import { Checkbox } from '@mui/material';

// Replace the hardcoded API_URL with a function to get the base URL
const getApiUrl = () => {
  return `${window.location.protocol}//${window.location.hostname}:8069`;
};

interface Video {
  id: string;
  path: string;
  size: number;
}

interface GridProps {
  videos: Video[];
  onVideoSelect: (videoPath: string, videoId: string, isSelected: boolean) => void;
  currentVideoId: string | null;
  onSortModelChange: (model: GridSortModel) => void;
  showCheckbox?: boolean;
}

function Grid({ videos, onVideoSelect, currentVideoId, onSortModelChange, showCheckbox = false }: GridProps) {
  const [rows, setRows] = useState<Video[]>([]);
  const queryClient = useQueryClient();

  useEffect(() => {
    console.log('Videos prop:', videos);
    setRows(videos);
  }, [videos]);

  const deleteMutation = useMutation(
    (id: string) => axios.delete(`${getApiUrl()}/api/v1/playlist/${id}`),
    {
      onSuccess: () => {
        queryClient.invalidateQueries('videos');
      },
    }
  );

  const formatSize = (params: GridRenderCellParams<Video, number>) => {
    const sizeInBytes = params.value;
    if (typeof sizeInBytes === 'number' && !isNaN(sizeInBytes)) {
      if (sizeInBytes >= 1073741824) {
        return `${(sizeInBytes / 1073741824).toFixed(2)} GB`;
      } else if (sizeInBytes >= 1048576) {
        return `${(sizeInBytes / 1048576).toFixed(2)} MB`;
      } else {
        return `${sizeInBytes} bytes`;
      }
    }
    return 'Unknown';
  };

  const columns: GridColDef[] = [
    showCheckbox ? {
      field: 'checkbox',
      headerName: '',
      width: 50,
      renderCell: (params: GridRenderCellParams<Video>) => (
        <Checkbox
          onChange={(event) => onVideoSelect(params.row.path, params.row.id, event.target.checked)}
        />
      ),
    } : null,
    { field: 'id', headerName: 'ID', width: 70 },
    { field: 'path', headerName: 'Name', width: 500 },
    {
      field: 'size',
      headerName: 'Size',
      width: 150,
      renderCell: formatSize,
      type: 'number',
    },
    {
      field: 'actions',
      headerName: 'Actions',
      width: 100,
      renderCell: (params: GridRenderCellParams<Video>) => (
        <IconButton onClick={() => deleteMutation.mutate(params.row.id)}>
          <DeleteIcon />
        </IconButton>
      ),
    },
  ].filter(Boolean) as GridColDef[];

  return (
    <DataGrid
      rows={rows}
      columns={columns}
      initialState={{
        pagination: {
          paginationModel: { pageSize: 5, page: 0 },
        },
      }}
      pageSizeOptions={[5, 10, 25, 50, 100, 250]}
      onRowClick={(params: GridRowParams) => onVideoSelect(params.row.path, params.row.id, false)}
      autoHeight
      onSortModelChange={onSortModelChange}
      getRowClassName={(params) => params.id === currentVideoId ? 'Mui-selected' : ''}
    />
  );
}

export default Grid;