import { useState, useEffect } from 'react';
import { DataGrid, GridColDef, GridRowParams, GridSortModel, GridRenderCellParams } from '@mui/x-data-grid';
import { IconButton } from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';
import { useMutation, useQueryClient } from 'react-query';
import axios from 'axios';

const API_URL = 'http://185.187.169.230:8069';

interface Video {
  id: string;
  path: string;
  size: number;
}

interface GridProps {
  videos: Video[];
  onVideoSelect: (videoPath: string, videoId: string) => void;
  currentVideoId: string | null;
}

function Grid({ videos, onVideoSelect, currentVideoId }: GridProps) {
  const [sortModel, setSortModel] = useState<GridSortModel>([]);
  const [rows, setRows] = useState<Video[]>([]);
  const queryClient = useQueryClient();

  useEffect(() => {
    console.log('Videos prop:', videos);
    setRows(videos);
  }, [videos]);

  const deleteMutation = useMutation(
    (id: string) => axios.delete(`${API_URL}/api/v1/playlist/${id}`),
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
      renderCell: (params) => (
        <IconButton onClick={() => deleteMutation.mutate(params.row.id)}>
          <DeleteIcon />
        </IconButton>
      ),
    },
  ];

  const handleSortModelChange = (newSortModel: GridSortModel) => {
    setSortModel(newSortModel);
  };

  return (
    <DataGrid
      rows={rows}
      columns={columns}
      initialState={{
        pagination: {
          paginationModel: { pageSize: 5, page: 0 },
        },
      }}
      pageSizeOptions={[5, 10, 25]}
      onRowClick={(params: GridRowParams) => onVideoSelect(`${API_URL}/videos/${params.row.path}`, params.row.id)}
      autoHeight
      sortModel={sortModel}
      onSortModelChange={handleSortModelChange}
      getRowClassName={(params) => params.id === currentVideoId ? 'Mui-selected' : ''}
    />
  );
}

export default Grid;