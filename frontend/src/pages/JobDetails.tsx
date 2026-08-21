import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useJobStore } from '../store/jobStore';
import { jobService } from '../api/jobs';
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { ProgressBar } from '../components/ui/ProgressBar';
import { Modal } from '../components/ui/Modal';
import { ArrowLeft, Download, XCircle, Trash2, Clock, Activity, AlertTriangle } from 'lucide-react';
import type { ExportResponse } from '../types/models';

export function JobDetails() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { progress, fetchProgress } = useJobStore();
  const jobProgress = id ? progress[id] : null;

  const [exportUrls, setExportUrls] = useState<{ json: string[], csv: string[] }>({ json: [], csv: [] });
  const [isLoadingExports, setIsLoadingExports] = useState(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);

  useEffect(() => {
    if (!id) return;

    fetchProgress(id);

    const isTerminal = jobProgress?.status === 'COMPLETED' || 
                       jobProgress?.status === 'FAILED' || 
                       jobProgress?.status === 'CANCELLED';

    if (isTerminal) return;

    const interval = setInterval(() => {
      fetchProgress(id);
    }, 1000);

    return () => clearInterval(interval);
  }, [id, fetchProgress, jobProgress?.status]);

  useEffect(() => {
    let mounted = true;
    if (jobProgress?.status === 'COMPLETED' && id) {
      if (!exportUrls.json.length && !exportUrls.csv.length && !isLoadingExports) {
        setIsLoadingExports(true);
        Promise.all([
          jobService.getExportURLs(id, 'json').catch(() => ({ status: 'error', urls: [] } as unknown as ExportResponse)),
          jobService.getExportURLs(id, 'csv').catch(() => ({ status: 'error', urls: [] } as unknown as ExportResponse))
        ]).then(([jsonRes, csvRes]) => {
          if (mounted) {
            setExportUrls({
              json: jsonRes.urls || [],
              csv: csvRes.urls || []
            });
            setIsLoadingExports(false);
          }
        });
      }
    }
    return () => { mounted = false; };
  }, [jobProgress?.status, id, exportUrls.json.length, exportUrls.csv.length, isLoadingExports]);

  const handleCancel = async () => {
    if (!id) return;
    try {
      await jobService.cancelJob(id);
      await fetchProgress(id);
    } catch (err) {
      console.error(err);
    }
  };

  const handleDelete = async () => {
    if (!id) return;
    try {
      await jobService.deleteJob(id);
      navigate('/');
    } catch (err) {
      console.error(err);
    }
  };

  if (!jobProgress) {
    return <div className="p-8 text-center text-muted-foreground animate-pulse">Loading job details...</div>;
  }

  const isRunning = jobProgress.status === 'RUNNING' || jobProgress.status === 'PENDING';

  return (
    <div className="max-w-5xl mx-auto space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="sm" onClick={() => navigate('/')}>
          <ArrowLeft className="h-4 w-4 mr-2" /> Back
        </Button>
        <h1 className="text-2xl font-bold truncate">Job {id}</h1>
        <Badge status={jobProgress.status} className="ml-auto" />
      </div>

      <Card>
        <CardHeader className="pb-4">
          <CardTitle>Processing Progress</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex justify-between items-end">
            <div>
              <div className="text-4xl font-bold tracking-tight">
                {jobProgress.percent_complete.toFixed(1)}%
              </div>
              <p className="text-sm text-muted-foreground mt-1">
                {jobProgress.processed_records.toLocaleString()} records processed
              </p>
            </div>
            {isRunning && (
              <div className="flex gap-2">
                <Button variant="danger" size="sm" onClick={handleCancel}>
                  <XCircle className="h-4 w-4 mr-2" /> Cancel Job
                </Button>
              </div>
            )}
          </div>

          <ProgressBar progress={jobProgress.percent_complete} />
          
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 pt-4 border-t border-border">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 rounded-lg">
                <Activity className="h-5 w-5" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Throughput</p>
                <p className="font-semibold">{Math.round(jobProgress.records_per_second).toLocaleString()} rec/s</p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <div className="p-2 bg-amber-100 dark:bg-amber-900/30 text-amber-600 rounded-lg">
                <AlertTriangle className="h-5 w-5" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Errors</p>
                <p className="font-semibold">{jobProgress.error_count.toLocaleString()}</p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <div className="p-2 bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 rounded-lg">
                <Clock className="h-5 w-5" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Start Time</p>
                <p className="font-semibold">
                  {new Date(jobProgress.start_time).toLocaleTimeString()}
                </p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {jobProgress.status === 'COMPLETED' && (
        <Card className="bg-linear-to-br from-indigo-50 to-white dark:from-indigo-950/20 dark:to-card border-indigo-100 dark:border-indigo-900/50">
          <CardHeader>
            <CardTitle className="text-indigo-900 dark:text-indigo-100">Exported Results</CardTitle>
          </CardHeader>
          <CardContent>
            {isLoadingExports ? (
              <p className="text-indigo-600 animate-pulse">Generating pre-signed S3 links...</p>
            ) : (
              <div className="space-y-4">
                <p className="text-sm text-muted-foreground">
                  Your data has been successfully processed and streamed to the ephemeral S3 bucket. Click below to download the final artifacts.
                </p>
                <div className="flex flex-wrap gap-4">
                  {exportUrls.json.map((url, i) => (
                    <Button key={i} onClick={() => window.open(url, '_blank')} className="bg-indigo-600 hover:bg-indigo-700">
                      <Download className="h-4 w-4 mr-2" /> JSON Result Part {i+1}
                    </Button>
                  ))}
                  {exportUrls.csv.map((url, i) => (
                    <Button key={i} variant="secondary" onClick={() => window.open(url, '_blank')}>
                      <Download className="h-4 w-4 mr-2" /> CSV Result Part {i+1}
                    </Button>
                  ))}
                  {exportUrls.json.length === 0 && exportUrls.csv.length === 0 && (
                     <p className="text-sm text-muted-foreground italic">No export targets configured or files found.</p>
                  )}
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      <div className="flex justify-end pt-8">
        <Button variant="ghost" className="text-red-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-950/30" onClick={() => setIsDeleteModalOpen(true)}>
          <Trash2 className="h-4 w-4 mr-2" /> Permanently Delete Job
        </Button>
      </div>

      <Modal
        isOpen={isDeleteModalOpen}
        onClose={() => setIsDeleteModalOpen(false)}
        title="Delete Job"
        footer={
          <>
            <Button variant="ghost" onClick={() => setIsDeleteModalOpen(false)}>Cancel</Button>
            <Button variant="danger" onClick={handleDelete}>Delete Permanently</Button>
          </>
        }
      >
        <p className="text-sm text-muted-foreground">
          Are you sure you want to permanently delete this job and its configuration? This action cannot be undone.
        </p>
      </Modal>
    </div>
  );
}
