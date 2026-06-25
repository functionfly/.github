import { useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  RefreshCw,
  Calendar,
  Clock,
  ToggleLeft,
  ToggleRight,
  Trash2,
  AlertCircle,
} from "lucide-react";
import { formatDistanceToNow, format } from "date-fns";
import { cn } from "@/lib/utils";
import {
  useRotationSchedules,
  useRotationSchedule,
  useSetAutoRotation,
  useCancelRotation,
} from "@/api/vault";
import { PlanGate } from "@/components/VaultEnterprise/PlanGate";
import { formatPlan, statusColor } from "@/components/VaultEnterprise/utils";
import type { VaultPlan, RotationSchedule } from "@/types/vault-enterprise";

interface RotationTabProps {
  plan: VaultPlan;
}

function RotationScheduleCard({ schedule }: { schedule: RotationSchedule }) {
  const setAutoRotation = useSetAutoRotation();
  const cancelRotation = useCancelRotation();

  const handleToggle = () => {
    setAutoRotation.mutate({
      secret_id: schedule.secret_id,
      enabled: !schedule.enabled,
      auto_rotate_interval: schedule.auto_rotate_interval,
      grace_period_hours: schedule.grace_period_hours,
      notify_stakeholders: schedule.notify_stakeholders,
      require_approval: schedule.require_approval,
    });
  };

  const handleCancel = () => {
    if (!confirm("Cancel this rotation schedule?")) return;
    cancelRotation.mutate({
      secretId: schedule.secret_id,
      body: { schedule_id: schedule.id, reason: "User cancelled" },
    });
  };

  return (
    <TableRow>
      <TableCell className="font-medium">{schedule.secret_name || schedule.secret_id.slice(0, 8)}</TableCell>
      <TableCell>
        <Badge variant="outline">{schedule.rotation_type}</Badge>
      </TableCell>
      <TableCell>
        {schedule.enabled ? (
          <span className="flex items-center gap-1 text-success text-sm">
            <ToggleRight className="h-4 w-4" />
            Active
          </span>
        ) : (
          <span className="flex items-center gap-1 text-muted-foreground text-sm">
            <ToggleLeft className="h-4 w-4" />
            Paused
          </span>
        )}
      </TableCell>
      <TableCell className="text-sm">
        {schedule.rotation_type === "automatic"
          ? `Every ${schedule.auto_rotate_interval} days`
          : schedule.scheduled_at
            ? format(new Date(schedule.scheduled_at), "PPp")
            : "—"}
      </TableCell>
      <TableCell className="text-sm">
        {schedule.next_rotation_at ? (
          <span title={new Date(schedule.next_rotation_at).toLocaleString()}>
            {formatDistanceToNow(new Date(schedule.next_rotation_at), { addSuffix: true })}
          </span>
        ) : (
          "—"
        )}
      </TableCell>
      <TableCell className="text-sm">
        {schedule.last_rotated_at ? (
          formatDistanceToNow(new Date(schedule.last_rotated_at), { addSuffix: true })
        ) : (
          "Never"
        )}
      </TableCell>
      <TableCell>
        <Badge variant={statusColor(schedule.status)}>{schedule.status}</Badge>
      </TableCell>
      <TableCell>
        <div className="flex gap-1">
          <Button
            size="sm"
            variant="outline"
            onClick={handleToggle}
            disabled={setAutoRotation.isPending}
            title={schedule.enabled ? "Pause rotation" : "Enable rotation"}
          >
            {schedule.enabled ? (
              <ToggleRight className="h-3 w-3" />
            ) : (
              <ToggleLeft className="h-3 w-3" />
            )}
          </Button>
          {schedule.status === "pending" && (
            <Button
              size="sm"
              variant="ghost"
              onClick={handleCancel}
              disabled={cancelRotation.isPending}
              title="Cancel scheduled rotation"
            >
              <Trash2 className="h-3 w-3 text-destructive" />
            </Button>
          )}
        </div>
      </TableCell>
    </TableRow>
  );
}

function RotationScheduleCardSkeleton() {
  return (
    <TableRow>
      <TableCell><Skeleton className="h-4 w-24" /></TableCell>
      <TableCell><Skeleton className="h-5 w-16" /></TableCell>
      <TableCell><Skeleton className="h-4 w-20" /></TableCell>
      <TableCell><Skeleton className="h-4 w-32" /></TableCell>
      <TableCell><Skeleton className="h-4 w-24" /></TableCell>
      <TableCell><Skeleton className="h-4 w-24" /></TableCell>
      <TableCell><Skeleton className="h-5 w-16" /></TableCell>
      <TableCell><Skeleton className="h-8 w-16" /></TableCell>
    </TableRow>
  );
}

export function RotationTab({ plan }: RotationTabProps) {
  const { data, isLoading } = useRotationSchedules();
  const schedules = data?.schedules ?? [];

  return (
    <PlanGate feature="rotationSchedules" plan={plan} title="Rotation Schedules" description="Manage automated and scheduled secret rotation.">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <RefreshCw className="h-5 w-5" />
            <CardTitle>Rotation Schedules</CardTitle>
          </div>
          <CardDescription>
            Configure automatic or scheduled rotation for your secrets.
            Rotation happens in the background and old values remain valid during the grace period.
            Plan: <strong>{formatPlan(plan)}</strong>
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Secret</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Schedule</TableHead>
                  <TableHead>Next Rotation</TableHead>
                  <TableHead>Last Rotated</TableHead>
                  <TableHead>State</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {Array.from({ length: 3 }).map((_, i) => (
                  <RotationScheduleCardSkeleton key={i} />
                ))}
              </TableBody>
            </Table>
          ) : schedules.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <RefreshCw className="h-12 w-12 mx-auto mb-4 opacity-20" />
              <p className="text-lg font-medium">No rotation schedules</p>
              <p className="text-sm mt-1">
                Open a secret&apos;s menu and select &quot;Rotate&quot; to set up rotation.
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Secret</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Schedule</TableHead>
                  <TableHead>Next Rotation</TableHead>
                  <TableHead>Last Rotated</TableHead>
                  <TableHead>State</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {schedules.map((schedule) => (
                  <RotationScheduleCard key={schedule.id} schedule={schedule} />
                ))}
              </TableBody>
            </Table>
          )}
          {schedules.length > 0 && (
            <p className="text-xs text-muted-foreground mt-4">
              {schedules.length} schedule{schedules.length !== 1 ? "s" : ""} configured
            </p>
          )}
        </CardContent>
      </Card>
    </PlanGate>
  );
}
