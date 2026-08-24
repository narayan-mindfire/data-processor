import { useEffect } from 'react';
import { useJobStore } from '../store/jobStore';
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { useNavigate } from 'react-router-dom';
import { PlusCircle, ArrowRight, Loader2, ChevronLeft, ChevronRight } from 'lucide-react';

export function Dashboard() {
  const { jobs, totalJobs, currentPage, pageSize, fetchJobs, isLoading } = useJobStore();
  const navigate = useNavigate();

  useEffect(() => {
    fetchJobs(1);
  }, [fetchJobs]);

  const jobsList = Object.values(jobs).sort(
    (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
  );

  const totalPages = Math.ceil(totalJobs / pageSize);

  return (
    <div className="space-y-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Pipeline Jobs</h1>
        <Button onClick={() => navigate('/jobs/new')}>
          <PlusCircle className="mr-2 h-4 w-4" />
          New Job
        </Button>
      </div>

      {isLoading ? (
        <div className="flex h-64 items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      ) : jobsList.length === 0 ? (
        <Card className="flex h-64 flex-col items-center justify-center space-y-4">
          <CardTitle className="text-xl">No jobs found</CardTitle>
          <p className="text-muted-foreground">You haven't created any pipeline jobs yet.</p>
          <Button onClick={() => navigate('/jobs/new')}>Create your first job</Button>
        </Card>
      ) : (
        <>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {jobsList.map((job) => (
              <Card key={job.id} className="hover:shadow-md transition-shadow cursor-pointer" onClick={() => navigate(`/jobs/${job.id}`)}>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium truncate pr-2">
                    {job.id}
                  </CardTitle>
                  <Badge status={job.status} />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{job.processed_records.toLocaleString()}</div>
                  <p className="text-xs text-muted-foreground mb-4">
                    {job.total_records > 0 
                      ? `Records processed out of ${job.total_records.toLocaleString()}`
                      : `Total records processed`}
                  </p>
                  <div className="flex justify-between items-center text-sm">
                    <span className="text-muted-foreground">
                      {new Date(job.created_at).toLocaleDateString()}
                    </span>
                    <Button variant="ghost" size="sm" className="h-8 w-8 p-0" onClick={(e) => { e.stopPropagation(); navigate(`/jobs/${job.id}`); }}>
                      <ArrowRight className="h-4 w-4" />
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
          
          <div className="flex items-center justify-between border-t border-border pt-4">
            <p className="text-sm text-muted-foreground">
              Showing <span className="font-medium">{(currentPage - 1) * pageSize + 1}</span> to <span className="font-medium">{Math.min(currentPage * pageSize, totalJobs)}</span> of <span className="font-medium">{totalJobs}</span> jobs
            </p>
            <div className="flex items-center gap-2">
              <Button 
                variant="secondary" 
                size="sm" 
                onClick={() => fetchJobs(currentPage - 1)}
                disabled={currentPage <= 1}
              >
                <ChevronLeft className="h-4 w-4 mr-1" /> Previous
              </Button>
              <Button 
                variant="secondary" 
                size="sm" 
                onClick={() => fetchJobs(currentPage + 1)}
                disabled={currentPage >= totalPages}
              >
                Next <ChevronRight className="h-4 w-4 ml-1" />
              </Button>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
