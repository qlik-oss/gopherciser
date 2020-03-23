# Using Gopherciser for pre-caching

Pre-caching is used to trigger certain selections and actions in an app in order to have the calculations cached. This can save valuable time as fetching something from the cache is much faster than performing the needed calculations when the first users enter the system.

During a restart of the engine all cached result sets are purged and during a reload, when the source data is modified, the result sets from previously cached calculations become invalid. So, to avoid poor performance for the first users after a restart, the unloading of an app, or a reload it may be beneficial to pre-cache common calculations.

Do the following to set up pre-caching using Gopherciser:

1. Create a pre-caching scenario.

   When creating the scenario, keep the following differences between pre-caching and load/performance scenarios in mind:
   * Pre-caching scenarios should not use randomization to the same extent as load/performance scenarios.
   * Usually only one concurrent user is needed to build up the cache.
   * Minimize the think times between actions as pre-caching scenarios do not have to replicate reality.
   * Use the SheetChanger action to cache the initial state of all objects on all sheets in an app.
  
   Save the pre-caching scenario as a script file (`.json` file).  

2. Create a script file (for example, a batch file, `.bat`, in case of Microsoft Windows, or a shell script file, `.sh`, in case of Linux) that executes the pre-caching scenario.

   Example of file contents (in case of a Microsoft Windows batch file):
   
   ```
   C:\performancetests\gopherciser\gopherciser execute -c C:\performancetests\gopherciser\Scenarios\MyPreCachingScript.json
   ```
  
3. Select a scheduling mechanism (for example, the Task Scheduler in Microsoft Windows or the cron job scheduler in Linux) and configure it to run the batch / shell script file when needed.
