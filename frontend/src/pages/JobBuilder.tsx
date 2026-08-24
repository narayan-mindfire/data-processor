import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { jobService } from '../api/jobs';
import type { JobConfig } from '../types/models';
import { Button } from '../components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { Select } from '../components/ui/Select';
import { useJobStore } from '../store/jobStore';
import { Plus, Trash2 } from 'lucide-react';

export function JobBuilder() {
  const navigate = useNavigate();
  const { fetchJobs } = useJobStore();
  
  const [config, setConfig] = useState<JobConfig>({
    sources: [{ url: '', type: 'json' }],
    validations: [],
    transformations: [],
    aggregations: [],
    export_targets: [{ type: 'database', target: 'job_exported_records' }],
    concurrency: { validation_workers: 10, transform_workers: 10 }
  });
  
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');

  const addSource = () => setConfig({ ...config, sources: [...config.sources, { url: '', type: 'json' }] });
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

  const addValidation = () => setConfig({ ...config, validations: [...(config.validations || []), { field: '', rule: 'not_empty' }] });
  const removeValidation = (index: number) => {
    const newVals = [...(config.validations || [])];
    newVals.splice(index, 1);
    setConfig({ ...config, validations: newVals });
  };
  const updateValidation = (index: number, field: string, value: string) => {
    const newVals = [...(config.validations || [])];
    newVals[index] = { ...newVals[index], [field]: value };
    setConfig({ ...config, validations: newVals });
  };

  const addTransformation = () => setConfig({ ...config, transformations: [...(config.transformations || []), { field: '', action: 'convert_to_float' }] });
  const removeTransformation = (index: number) => {
    const newTrans = [...(config.transformations || [])];
    newTrans.splice(index, 1);
    setConfig({ ...config, transformations: newTrans });
  };
  const updateTransformation = (index: number, field: string, value: string) => {
    const newTrans = [...(config.transformations || [])];
    newTrans[index] = { ...newTrans[index], [field]: value };
    setConfig({ ...config, transformations: newTrans });
  };

  const addAggregation = () => setConfig({ ...config, aggregations: [...(config.aggregations || []), { type: 'count', field: '', output_name: '' }] });
  const removeAggregation = (index: number) => {
    const newAgg = [...(config.aggregations || [])];
    newAgg.splice(index, 1);
    setConfig({ ...config, aggregations: newAgg });
  };
  const updateAggregation = (index: number, field: string, value: string) => {
    const newAgg = [...(config.aggregations || [])];
    newAgg[index] = { ...newAgg[index], [field]: value };
    setConfig({ ...config, aggregations: newAgg });
  };

  const addExportTarget = () => setConfig({ ...config, export_targets: [...(config.export_targets || []), { type: 'database', target: '' }] });
  const removeExportTarget = (index: number) => {
    const newExp = [...(config.export_targets || [])];
    newExp.splice(index, 1);
    setConfig({ ...config, export_targets: newExp });
  };
  const updateExportTarget = (index: number, field: string, value: string) => {
    const newExp = [...(config.export_targets || [])];
    newExp[index] = { ...newExp[index], [field]: value };
    setConfig({ ...config, export_targets: newExp });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsSubmitting(true);
    
    try {
      const response = await jobService.createJob(config);
      await fetchJobs();
      navigate(`/jobs/${response.id}`);
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || 'Failed to create job');
      setIsSubmitting(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto space-y-6 pb-20">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Create Pipeline Job</h1>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        
        {/* SOURCES */}
        <Card>
          <CardHeader>
            <CardTitle className="text-xl flex items-center justify-between">
              Data Sources
              <Button type="button" variant="ghost" size="sm" onClick={addSource}>
                <Plus className="h-4 w-4 mr-2" /> Add Source
              </Button>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {config.sources.map((source, idx) => (
              <div key={idx} className="flex gap-4 items-start bg-muted/50 p-4 rounded-lg border border-border flex-wrap md:flex-nowrap">
                <div className="w-full md:flex-1 space-y-2">
                  <label className="text-sm font-medium">Source URL</label>
                  <Input 
                    placeholder="https://example.com/data.json"
                    value={source.url}
                    onChange={(e) => updateSource(idx, 'url', e.target.value)}
                    required
                  />
                </div>
                <div className="w-full md:w-1/4 space-y-2">
                  <label className="text-sm font-medium">Format</label>
                  <Select
                    value={source.type}
                    onChange={(e) => updateSource(idx, 'type', e.target.value)}
                  >
                    <option value="json">JSON</option>
                    <option value="csv">CSV</option>
                  </Select>
                </div>
                {source.type === 'json' && (
                  <div className="w-full md:flex-1 space-y-2">
                    <label className="text-sm font-medium text-muted-foreground">JSON Array Path</label>
                    <Input 
                      placeholder="e.g. results"
                      value={source.json_array_path || ''}
                      onChange={(e) => updateSource(idx, 'json_array_path', e.target.value)}
                    />
                  </div>
                )}
                <div className="pt-7">
                  <Button type="button" variant="danger" size="sm" onClick={() => removeSource(idx)} disabled={config.sources.length === 1}>
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            ))}
          </CardContent>
        </Card>

        {/* VALIDATIONS */}
        <Card>
          <CardHeader>
            <CardTitle className="text-xl flex items-center justify-between">
              Validations
              <Button type="button" variant="ghost" size="sm" onClick={addValidation}>
                <Plus className="h-4 w-4 mr-2" /> Add Validation
              </Button>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {config.validations?.map((val, idx) => (
              <div key={idx} className="flex gap-4 items-start bg-muted/50 p-4 rounded-lg border border-border">
                <div className="flex-1 space-y-2">
                  <label className="text-sm font-medium">Field Name</label>
                  <Input value={val.field} onChange={(e) => updateValidation(idx, 'field', e.target.value)} required />
                </div>
                <div className="flex-1 space-y-2">
                  <label className="text-sm font-medium">Rule</label>
                  <Select value={val.rule} onChange={(e) => updateValidation(idx, 'rule', e.target.value)}>
                    <option value="not_empty">Not Empty</option>
                    <option value="is_numeric">Is Numeric</option>
                  </Select>
                </div>
                <div className="pt-7">
                  <Button type="button" variant="danger" size="sm" onClick={() => removeValidation(idx)}>
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            ))}
            {(!config.validations || config.validations.length === 0) && (
              <p className="text-sm text-muted-foreground italic">No validations configured.</p>
            )}
          </CardContent>
        </Card>

        {/* TRANSFORMATIONS */}
        <Card>
          <CardHeader>
            <CardTitle className="text-xl flex items-center justify-between">
              Transformations
              <Button type="button" variant="ghost" size="sm" onClick={addTransformation}>
                <Plus className="h-4 w-4 mr-2" /> Add Transformation
              </Button>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {config.transformations?.map((trans, idx) => (
              <div key={idx} className="flex gap-4 items-start bg-muted/50 p-4 rounded-lg border border-border">
                <div className="flex-1 space-y-2">
                  <label className="text-sm font-medium">Field Name</label>
                  <Input value={trans.field} onChange={(e) => updateTransformation(idx, 'field', e.target.value)} required />
                </div>
                <div className="flex-1 space-y-2">
                  <label className="text-sm font-medium">Action</label>
                  <Select value={trans.action} onChange={(e) => updateTransformation(idx, 'action', e.target.value)}>
                    <option value="convert_to_float">Convert to Float</option>
                    <option value="convert_to_int">Convert to Int</option>
                    <option value="fill_empty">Fill Empty</option>
                  </Select>
                </div>
                {trans.action === 'fill_empty' && (
                  <div className="flex-1 space-y-2">
                    <label className="text-sm font-medium">Default Value</label>
                    <Input value={trans.default_value || ''} onChange={(e) => updateTransformation(idx, 'default_value', e.target.value)} required />
                  </div>
                )}
                <div className="pt-7">
                  <Button type="button" variant="danger" size="sm" onClick={() => removeTransformation(idx)}>
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            ))}
            {(!config.transformations || config.transformations.length === 0) && (
              <p className="text-sm text-muted-foreground italic">No transformations configured.</p>
            )}
          </CardContent>
        </Card>

        {/* AGGREGATIONS */}
        <Card>
          <CardHeader>
            <CardTitle className="text-xl flex items-center justify-between">
              Aggregations
              <Button type="button" variant="ghost" size="sm" onClick={addAggregation}>
                <Plus className="h-4 w-4 mr-2" /> Add Aggregation
              </Button>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {config.aggregations?.map((agg, idx) => (
              <div key={idx} className="flex gap-4 items-start bg-muted/50 p-4 rounded-lg border border-border">
                <div className="w-1/4 space-y-2">
                  <label className="text-sm font-medium">Type</label>
                  <Select value={agg.type} onChange={(e) => updateAggregation(idx, 'type', e.target.value)}>
                    <option value="count">Count</option>
                    <option value="sum">Sum</option>
                    <option value="average">Average</option>
                  </Select>
                </div>
                <div className="flex-1 space-y-2">
                  <label className="text-sm font-medium">Target Field</label>
                  <Input value={agg.field} onChange={(e) => updateAggregation(idx, 'field', e.target.value)} required />
                </div>
                <div className="flex-1 space-y-2">
                  <label className="text-sm font-medium">Output Name</label>
                  <Input value={agg.output_name} onChange={(e) => updateAggregation(idx, 'output_name', e.target.value)} required />
                </div>
                <div className="pt-7">
                  <Button type="button" variant="danger" size="sm" onClick={() => removeAggregation(idx)}>
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            ))}
            {(!config.aggregations || config.aggregations.length === 0) && (
              <p className="text-sm text-muted-foreground italic">No aggregations configured.</p>
            )}
          </CardContent>
        </Card>

        {/* EXPORT TARGETS */}
        <Card>
          <CardHeader>
            <CardTitle className="text-xl flex items-center justify-between">
              Export Targets
              <Button type="button" variant="ghost" size="sm" onClick={addExportTarget}>
                <Plus className="h-4 w-4 mr-2" /> Add Export Target
              </Button>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {config.export_targets?.map((exp, idx) => (
              <div key={idx} className="flex gap-4 items-start bg-muted/50 p-4 rounded-lg border border-border">
                <div className="w-1/3 space-y-2">
                  <label className="text-sm font-medium">Type</label>
                  <Select value={exp.type} onChange={(e) => updateExportTarget(idx, 'type', e.target.value)}>
                    <option value="database">Database</option>
                  </Select>
                </div>
                <div className="flex-1 space-y-2">
                  <label className="text-sm font-medium">Target Table/Location</label>
                  <Input value={exp.target || ''} onChange={(e) => updateExportTarget(idx, 'target', e.target.value)} required />
                </div>
                <div className="pt-7">
                  <Button type="button" variant="danger" size="sm" onClick={() => removeExportTarget(idx)}>
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            ))}
            {(!config.export_targets || config.export_targets.length === 0) && (
              <p className="text-sm text-muted-foreground italic">No export targets configured.</p>
            )}
          </CardContent>
        </Card>

        {/* CONCURRENCY */}
        <Card>
          <CardHeader>
            <CardTitle className="text-xl">Concurrency Settings</CardTitle>
          </CardHeader>
          <CardContent className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Validation Workers</label>
              <Input 
                type="number"
                min="1"
                max="100"
                value={config.concurrency?.validation_workers}
                onChange={(e) => setConfig({
                  ...config, 
                  concurrency: { ...config.concurrency!, validation_workers: parseInt(e.target.value) || 1 }
                })}
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Transform Workers</label>
              <Input 
                type="number"
                min="1"
                max="100"
                value={config.concurrency?.transform_workers}
                onChange={(e) => setConfig({
                  ...config, 
                  concurrency: { ...config.concurrency!, transform_workers: parseInt(e.target.value) || 1 }
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

        <div className="flex justify-end gap-4 sticky bottom-4 p-4 bg-background/80 backdrop-blur-sm rounded-lg border border-border z-10 shadow-lg">
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
