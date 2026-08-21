import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { jobService } from '../api/jobs';
import { JobConfig } from '../types/models';
import { Button } from '../components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { Select } from '../components/ui/Select';
import { useJobStore } from '../store/jobStore';
import { Plus, Trash2, Code } from 'lucide-react';

export function JobBuilder() {
  const navigate = useNavigate();
  const { fetchJobs } = useJobStore();
  
  const [config, setConfig] = useState<JobConfig>({
    sources: [{ url: '', format: 'json' }],
    concurrency: { max_workers: 5, batch_size: 1000 }
  });
  
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');

  const addSource = () => {
    setConfig({
      ...config,
      sources: [...config.sources, { url: '', format: 'json' }]
    });
  };

  const removeSource = (index: number) => {
    const newSources = [...config.sources];
    newSources.splice(index, 1);
    setConfig({ ...config, sources: newSources });
  };

  const updateSource = (index: number, field: string, value: string) => {
    const newSources = [...config.sources];
    newSources[index] = { ...newSources[index], [field]: value };
    setConfig({ ...config, sources: newSources });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsSubmitting(true);
    
    try {
      const response = await jobService.createJob(config);
      await fetchJobs();
      navigate(`/jobs/${response.job_id}`);
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || 'Failed to create job');
      setIsSubmitting(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Create Pipeline Job</h1>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle className="text-xl flex items-center justify-between">
              Data Sources
              <Button type="button" variant="ghost" size="sm" onClick={addSource}>
                <Plus className="h-4 w-4 mr-2" />
                Add Source
              </Button>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {config.sources.map((source, idx) => (
              <div key={idx} className="flex gap-4 items-start bg-muted/50 p-4 rounded-lg border border-border">
                <div className="flex-1 space-y-2">
                  <label className="text-sm font-medium">Source URL</label>
                  <Input 
                    placeholder="https://example.com/data.json"
                    value={source.url}
                    onChange={(e) => updateSource(idx, 'url', e.target.value)}
                    required
                  />
                </div>
                <div className="w-1/4 space-y-2">
                  <label className="text-sm font-medium">Format</label>
                  <Select
                    value={source.format}
                    onChange={(e) => updateSource(idx, 'format', e.target.value)}
                  >
                    <option value="json">JSON</option>
                    <option value="csv">CSV</option>
                  </Select>
                </div>
                {config.sources.length > 1 && (
                  <div className="pt-7">
                    <Button type="button" variant="danger" size="sm" onClick={() => removeSource(idx)}>
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                )}
              </div>
            ))}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-xl">Concurrency Settings</CardTitle>
          </CardHeader>
          <CardContent className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Max Workers</label>
              <Input 
                type="number"
                min="1"
                max="100"
                value={config.concurrency?.max_workers}
                onChange={(e) => setConfig({
                  ...config, 
                  concurrency: { ...config.concurrency!, max_workers: parseInt(e.target.value) || 1 }
                })}
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Batch Size</label>
              <Input 
                type="number"
                min="100"
                max="10000"
                value={config.concurrency?.batch_size}
                onChange={(e) => setConfig({
                  ...config, 
                  concurrency: { ...config.concurrency!, batch_size: parseInt(e.target.value) || 1000 }
                })}
              />
            </div>
          </CardContent>
        </Card>

        {error && (
          <div className="p-4 bg-red-100 text-red-900 border border-red-200 rounded-md">
            {error}
          </div>
        )}

        <div className="flex justify-end gap-4">
          <Button type="button" variant="ghost" onClick={() => navigate('/')}>
            Cancel
          </Button>
          <Button type="submit" isLoading={isSubmitting}>
            Start Pipeline
          </Button>
        </div>
      </form>
    </div>
  );
}
